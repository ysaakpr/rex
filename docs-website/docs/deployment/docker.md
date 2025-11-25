# Docker Deployment

Deploy Rex using Docker and Docker Compose.

## Development Deployment

### Docker Compose Setup

The included `docker-compose.yml` runs all services locally:

```yaml
services:
  - postgres (database)
  - redis (cache & queue)
  - supertokens-db (SuperTokens database)
  - supertokens (authentication service)
  - api (Go backend)
  - worker (background jobs)
  - frontend (React app)
  - mailhog (email testing)
```

### Quick Start

```bash
# Clone repository
git clone https://github.com/yourorg/utm-backend
cd utm-backend

# Configure environment
cp .env.example .env

# Start all services
docker-compose up -d

# Run migrations
docker-compose exec api /app/migrate up

# Check status
docker-compose ps
```

### Service URLs

| Service | URL | Description |
|---------|-----|-------------|
| Frontend | http://localhost:3000 | React application |
| API | http://localhost:8080 | Backend REST API |
| MailHog | http://localhost:8025 | Email viewer |
| PostgreSQL | localhost:5432 | Database |
| Redis | localhost:6379 | Cache/Queue |

## Production Deployment

### Prerequisites

- Docker Engine 20.10+
- Docker Compose V2
- Domain name with DNS configured
- SSL certificates (Let's Encrypt recommended)

### Production docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    networks:
      - utm-network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped
    networks:
      - utm-network

  supertokens-db:
    image: postgres:15
    environment:
      POSTGRES_USER: supertokens
      POSTGRES_PASSWORD: ${SUPERTOKENS_DB_PASSWORD}
      POSTGRES_DB: supertokens
    volumes:
      - supertokens_db_data:/var/lib/postgresql/data
    restart: unless-stopped
    networks:
      - utm-network

  supertokens:
    image: registry.supertokens.io/supertokens/supertokens-postgresql:latest
    environment:
      POSTGRESQL_USER: supertokens
      POSTGRESQL_PASSWORD: ${SUPERTOKENS_DB_PASSWORD}
      POSTGRESQL_DATABASE_NAME: supertokens
      POSTGRESQL_HOST: supertokens-db
      POSTGRESQL_PORT: 5432
    depends_on:
      - supertokens-db
    restart: unless-stopped
    networks:
      - utm-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3567/hello"]
      interval: 30s
      timeout: 10s
      retries: 5

  api:
    image: yourregistry/utm-backend-api:latest
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - SUPERTOKENS_CONNECTION_URI=http://supertokens:3567
      - SUPERTOKENS_API_KEY=${SUPERTOKENS_API_KEY}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - API_DOMAIN=https://api.yourdomain.com
      - WEBSITE_DOMAIN=https://yourdomain.com
    depends_on:
      - postgres
      - redis
      - supertokens
    restart: unless-stopped
    networks:
      - utm-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  worker:
    image: yourregistry/utm-backend-worker:latest
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=${SMTP_PORT}
      - SMTP_USER=${SMTP_USER}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    networks:
      - utm-network

  frontend:
    image: yourregistry/utm-backend-frontend:latest
    restart: unless-stopped
    networks:
      - utm-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - certbot_data:/var/www/certbot:ro
    depends_on:
      - api
      - frontend
    restart: unless-stopped
    networks:
      - utm-network

volumes:
  postgres_data:
  redis_data:
  supertokens_db_data:
  certbot_data:

networks:
  utm-network:
    driver: bridge
```

### Nginx Configuration

**nginx.conf**:
```nginx
events {
    worker_connections 1024;
}

http {
    upstream api_backend {
        server api:8080;
    }

    upstream frontend_backend {
        server frontend:3000;
    }

    # Redirect HTTP to HTTPS
    server {
        listen 80;
        server_name yourdomain.com www.yourdomain.com api.yourdomain.com;
        
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }
        
        location / {
            return 301 https://$host$request_uri;
        }
    }

    # Frontend (HTTPS)
    server {
        listen 443 ssl http2;
        server_name yourdomain.com www.yourdomain.com;

        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        location / {
            proxy_pass http://frontend_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # API (HTTPS)
    server {
        listen 443 ssl http2;
        server_name api.yourdomain.com;

        ssl_certificate /etc/nginx/ssl/fullchain.pem;
        ssl_certificate_key /etc/nginx/ssl/privkey.pem;
        
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        location / {
            proxy_pass http://api_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # WebSocket support (if needed)
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
        }
    }
}
```

## Building Images

### API Image

**Dockerfile**:
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./cmd/migrate

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrate .
COPY migrations ./migrations

EXPOSE 8080

CMD ["./api"]
```

**Build**:
```bash
docker build -t yourregistry/utm-backend-api:latest .
docker push yourregistry/utm-backend-api:latest
```

### Frontend Image

**Dockerfile**:
```dockerfile
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Runtime stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

## SSL/TLS with Let's Encrypt

### Using Certbot

1. **Install Certbot**:
   ```bash
   docker run -it --rm \
     -v "${PWD}/ssl:/etc/letsencrypt" \
     -v "${PWD}/certbot:/var/www/certbot" \
     certbot/certbot certonly \
     --webroot \
     --webroot-path=/var/www/certbot \
     -d yourdomain.com \
     -d www.yourdomain.com \
     -d api.yourdomain.com \
     --email your@email.com \
     --agree-tos \
     --no-eff-email
   ```

2. **Auto-Renewal**:
   ```bash
   # Add to crontab
   0 3 * * * docker run --rm \
     -v "${PWD}/ssl:/etc/letsencrypt" \
     -v "${PWD}/certbot:/var/www/certbot" \
     certbot/certbot renew \
     && docker-compose exec nginx nginx -s reload
   ```

## Deployment Steps

1. **Prepare Server**:
   ```bash
   # Update system
   sudo apt update && sudo apt upgrade -y

   # Install Docker
   curl -fsSL https://get.docker.com | sh

   # Install Docker Compose
   sudo apt install docker-compose-plugin
   ```

2. **Clone Repository**:
   ```bash
   git clone https://github.com/yourorg/utm-backend
   cd utm-backend
   ```

3. **Configure Environment**:
   ```bash
   cp .env.example .env
   nano .env  # Edit with production values
   ```

4. **Build Images** (or pull from registry):
   ```bash
   # Option 1: Build locally
   docker-compose -f docker-compose.prod.yml build

   # Option 2: Pull from registry
   docker-compose -f docker-compose.prod.yml pull
   ```

5. **Start Services**:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

6. **Run Migrations**:
   ```bash
   docker-compose exec api ./migrate up
   ```

7. **Create Platform Admin**:
   ```bash
   ./scripts/create-platform-admin.sh admin@yourdomain.com
   ```

8. **Verify**:
   ```bash
   curl https://api.yourdomain.com/health
   # Should return: {"status": "ok"}
   ```

## Monitoring

### Health Checks

```bash
# Check all services
docker-compose ps

# Check logs
docker-compose logs -f api
docker-compose logs -f worker

# Check resource usage
docker stats
```

### Log Management

**Using Docker Logging**:
```yaml
services:
  api:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

**Centralized Logging** (optional):
```yaml
services:
  api:
    logging:
      driver: "syslog"
      options:
        syslog-address: "tcp://logserver:514"
```

## Backup Strategy

### Database Backup

```bash
# Backup
docker-compose exec postgres pg_dump -U ${DB_USER} ${DB_NAME} | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz

# Restore
gunzip < backup.sql.gz | docker-compose exec -T postgres psql -U ${DB_USER} ${DB_NAME}
```

### Automated Backups

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"

# Backup database
docker-compose exec postgres pg_dump -U $DB_USER $DB_NAME | gzip > ${BACKUP_DIR}/db_${DATE}.sql.gz

# Backup Redis (if persistent)
docker-compose exec redis redis-cli BGSAVE

# Backup volumes
tar -czf ${BACKUP_DIR}/volumes_${DATE}.tar.gz /var/lib/docker/volumes/utm-backend_*

# Keep only last 7 days
find ${BACKUP_DIR} -name "*.gz" -mtime +7 -delete
```

**Cron Job**:
```bash
0 2 * * * /path/to/backup.sh
```

## Scaling

### Horizontal Scaling

Run multiple API instances:

```yaml
services:
  api:
    deploy:
      replicas: 3
```

Use load balancer:
```nginx
upstream api_backend {
    server api1:8080;
    server api2:8080;
    server api3:8080;
}
```

### Worker Scaling

Run multiple workers:

```yaml
services:
  worker:
    deploy:
      replicas: 2
```

## Next Steps

- **[Environment Variables](/deployment/environment)** - Configure all settings
- **[Database Migrations](/deployment/migrations)** - Manage schema changes
- **[First Admin Setup](/deployment/first-admin)** - Create platform admin

