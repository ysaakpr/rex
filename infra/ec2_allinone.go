package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AllInOneResources struct {
	Instance          *ec2.Instance
	SecurityGroup     *ec2.SecurityGroup
	PublicIP          pulumi.StringOutput
	PrivateIP         pulumi.StringOutput
	PublicDNS         pulumi.StringOutput
	MasterUsername    string
	MainDBName        string
	SuperTokensDBName string
	MasterPassword    pulumi.StringOutput
}

func createAllInOneEC2(ctx *pulumi.Context, projectName, environment string, network *NetworkingResources,
	masterUsername string, masterPassword, supertokensApiKey pulumi.StringOutput,
	repositories *ECRRepositories, tags pulumi.StringMap) (*AllInOneResources, error) {

	mainDBName := "rex_backend"
	supertokensDBName := "supertokens"

	// Create Security Group for the all-in-one EC2 instance
	ec2SG, err := ec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-%s-allinone-sg", projectName, environment), &ec2.SecurityGroupArgs{
		VpcId:       network.VpcID,
		Description: pulumi.String("Security group for all-in-one EC2 instance (Everything in Docker Compose)"),
		Ingress: ec2.SecurityGroupIngressArray{
			// HTTP port (nginx reverse proxy)
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(80),
				ToPort:      pulumi.Int(80),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("HTTP access (nginx reverse proxy for /api and /auth)"),
			},
			// HTTPS port (optional, for future SSL)
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(443),
				ToPort:      pulumi.Int(443),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("HTTPS access"),
			},
			// SSH access (optional, for debugging - can be removed if using SSM)
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(22),
				ToPort:      pulumi.Int(22),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("SSH access"),
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-allinone-sg", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Create IAM role for EC2 instance (for Systems Manager and ECR access)
	ec2Role, err := iam.NewRole(ctx, fmt.Sprintf("%s-%s-allinone-role", projectName, environment), &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {
					"Service": "ec2.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}]
		}`),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-allinone-role", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Attach SSM policy for Systems Manager access
	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-%s-allinone-ssm-policy", projectName, environment), &iam.RolePolicyAttachmentArgs{
		Role:      ec2Role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
	})
	if err != nil {
		return nil, err
	}

	// Attach ECR read policy for pulling Docker images
	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-%s-allinone-ecr-policy", projectName, environment), &iam.RolePolicyAttachmentArgs{
		Role:      ec2Role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"),
	})
	if err != nil {
		return nil, err
	}

	// Create instance profile
	instanceProfile, err := iam.NewInstanceProfile(ctx, fmt.Sprintf("%s-%s-allinone-profile", projectName, environment), &iam.InstanceProfileArgs{
		Role: ec2Role.Name,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-allinone-profile", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	// Get the latest Ubuntu 22.04 LTS AMI for ARM64 (Graviton)
	ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		MostRecent: pulumi.BoolRef(true),
		Owners:     []string{"099720109477"}, // Canonical
		Filters: []ec2.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-arm64-server-*"},
			},
			{
				Name:   "virtualization-type",
				Values: []string{"hvm"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// User data script to install Docker, Docker Compose, and start all services
	// Using gzip compression to stay under AWS 16KB limit
	userData := pulumi.All(masterPassword, pulumi.String(masterUsername), pulumi.String(mainDBName),
		pulumi.String(supertokensDBName), supertokensApiKey, repositories.APIRepoURL, repositories.WorkerRepoURL).
		ApplyT(func(args []interface{}) (string, error) {
			password := args[0].(string)
			username := args[1].(string)
			mainDB := args[2].(string)
			supertokensDB := args[3].(string)
			apiKey := args[4].(string)
			apiRepo := args[5].(string)
			workerRepo := args[6].(string)

			scriptTemplate := `#!/bin/bash
set -e

# Update system
apt-get update
apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
systemctl enable docker
systemctl start docker

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Install AWS CLI for ECR login
apt-get install -y awscli jq

# Login to ECR
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin $(echo %s | cut -d'/' -f1)

# Create application directory
mkdir -p /app
cd /app

# Create docker-compose.yml
cat > docker-compose.yml <<'COMPOSE_EOF'
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    container_name: rex-postgres
    environment:
      POSTGRES_USER: %s
      POSTGRES_PASSWORD: %s
      POSTGRES_DB: %s
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - app-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U %s"]
      interval: 10s
      timeout: 5s
      retries: 5
    command: postgres -c 'max_connections=200'

  redis:
    image: redis:7-alpine
    container_name: rex-redis
    command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    networks:
      - app-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  supertokens:
    image: registry.supertokens.io/supertokens/supertokens-postgresql:7.0
    container_name: rex-supertokens
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      POSTGRESQL_CONNECTION_URI: postgresql://%s:%s@postgres:5432/%s
      API_KEYS: %s
    networks:
      - app-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3567/hello"]
      interval: 30s
      timeout: 10s
      retries: 3

  api:
    image: %s:latest
    container_name: rex-api
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      supertokens:
        condition: service_healthy
    environment:
      APP_ENV: dev
      APP_PORT: "8080"
      DB_HOST: postgres
      DB_PORT: "5432"
      DB_USER: %s
      DB_PASSWORD: %s
      DB_NAME: %s
      DB_SSL_MODE: disable
      REDIS_HOST: redis
      REDIS_PORT: "6379"
      SUPERTOKENS_CONNECTION_URI: http://supertokens:3567
      SUPERTOKENS_API_KEY: %s
      LOG_LEVEL: info
      LOG_FORMAT: json
    networks:
      - app-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  worker:
    image: %s:latest
    container_name: rex-worker
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      APP_ENV: dev
      DB_HOST: postgres
      DB_PORT: "5432"
      DB_USER: %s
      DB_PASSWORD: %s
      DB_NAME: %s
      DB_SSL_MODE: disable
      REDIS_HOST: redis
      REDIS_PORT: "6379"
      LOG_LEVEL: info
      LOG_FORMAT: json
    networks:
      - app-network
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    container_name: rex-nginx
    depends_on:
      api:
        condition: service_healthy
      supertokens:
        condition: service_healthy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - certbot-conf:/etc/letsencrypt
      - certbot-www:/var/www/certbot
    networks:
      - app-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:80/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    command: "/bin/sh -c 'while :; do sleep 6h & wait $${!}; nginx -s reload; done & nginx -g \"daemon off;\"'"

  certbot:
    image: certbot/certbot:latest
    container_name: rex-certbot
    volumes:
      - certbot-conf:/etc/letsencrypt
      - certbot-www:/var/www/certbot
    entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew; sleep 12h & wait $${!}; done;'"
    networks:
      - app-network
    restart: unless-stopped

networks:
  app-network:
    driver: bridge

volumes:
  postgres-data:
  redis-data:
  certbot-conf:
  certbot-www:
COMPOSE_EOF

# Create nginx configuration with HTTPS support
cat > nginx.conf <<'NGINX_EOF'
events {
    worker_connections 1024;
}

http {
    # Basic settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 20M;

    # Logging
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # Upstream definitions (internal HTTP communication)
    upstream api_backend {
        server api:8080;
    }

    upstream supertokens_backend {
        server supertokens:3567;
    }

    # HTTP Server - Redirect to HTTPS (except for Let's Encrypt challenges)
    server {
        listen 80;
        server_name _;

        # Let's Encrypt ACME challenge
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        # Health check (allow HTTP for monitoring)
        location /health {
            proxy_pass http://api_backend/health;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            access_log off;
        }

        # Redirect all other HTTP traffic to HTTPS
        location / {
            return 301 https://$host$request_uri;
        }
    }

    # HTTPS Server - Main application
    server {
        listen 443 ssl http2;
        server_name _;

        # SSL Configuration
        ssl_certificate /etc/letsencrypt/live/DOMAIN_PLACEHOLDER/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/DOMAIN_PLACEHOLDER/privkey.pem;
        
        # Fallback to self-signed if Let's Encrypt not available
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        # SSL Security Settings
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;

        # Security Headers
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;

        # API routes - proxy to Go API (internal HTTP)
        location /api/ {
            proxy_pass http://api_backend/api/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
            proxy_cache_bypass $http_upgrade;
        }

        # SuperTokens routes - proxy to SuperTokens (internal HTTP)
        location /auth/ {
            proxy_pass http://supertokens_backend/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
            proxy_cache_bypass $http_upgrade;
        }

        # Health check endpoint
        location /health {
            proxy_pass http://api_backend/health;
            proxy_http_version 1.1;
            proxy_set_header Host $host;
            access_log off;
        }

        # Root endpoint
        location / {
            return 200 'Rex Backend - All-in-One Mode (HTTPS)\n\nAvailable endpoints:\n- /api/* - Backend API\n- /auth/* - SuperTokens Authentication\n- /health - Health check\n';
            add_header Content-Type text/plain;
        }
    }
}
NGINX_EOF

# Create SSL setup script
cat > setup-ssl.sh <<'SSL_EOF'
#!/bin/bash
set -e

echo "=== SSL Certificate Setup ==="

# Get the public IP (Elastic IP - static across restarts)
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
echo "Public IP (Elastic IP): $PUBLIC_IP"

# Create self-signed certificate directory
mkdir -p /app/ssl

# Generate self-signed certificate for IP address
if [ ! -f /app/ssl/cert.pem ]; then
    echo "Generating self-signed certificate for IP: $PUBLIC_IP..."
    
    # Create OpenSSL config for IP-based certificate
    cat > /tmp/openssl-ip.conf <<SSLCONF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
C=US
ST=State
L=City
O=Organization
CN=$PUBLIC_IP

[v3_req]
subjectAltName = @alt_names

[alt_names]
IP.1 = $PUBLIC_IP
SSLCONF

    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /app/ssl/key.pem \
        -out /app/ssl/cert.pem \
        -config /tmp/openssl-ip.conf \
        -extensions v3_req
    
    echo "Self-signed certificate generated for IP: $PUBLIC_IP"
fi

# Update nginx.conf with IP address
sed -i "s/DOMAIN_PLACEHOLDER/$PUBLIC_IP/g" /app/nginx.conf

# Restart nginx to apply changes
docker-compose restart nginx

echo "=== SSL Setup Complete ==="
echo ""
echo "✓ Self-signed certificate active for IP: $PUBLIC_IP"
echo "✓ Elastic IP persists across instance restarts/replacements"
echo ""
echo "Access your instance:"
echo "- HTTP: http://$PUBLIC_IP (redirects to HTTPS)"
echo "- HTTPS: https://$PUBLIC_IP (self-signed, browser will warn)"
echo ""
echo "=== Let's Encrypt Certificate Options ==="
echo ""
echo "Option 1: Let's Encrypt with IP (Supported since 2024)"
echo "  Requirements:"
echo "    - Public IP (✓ You have Elastic IP)"
echo "    - Must use standalone mode (certbot stops nginx temporarily)"
echo ""
echo "  Command:"
echo "    docker-compose stop nginx"
echo "    docker-compose run --rm -p 80:80 certbot certonly --standalone \\"
echo "      --preferred-challenges http \\"
echo "      --email your-email@example.com \\"
echo "      --agree-tos \\"
echo "      --no-eff-email \\"
echo "      -d $PUBLIC_IP"
echo "    docker-compose start nginx"
echo ""
echo "Option 2: Let's Encrypt with Domain (Recommended)"
echo "  1. Point a domain to this IP: $PUBLIC_IP"
echo "  2. Wait for DNS propagation (1-2 minutes)"
echo "  3. Request certificate:"
echo "     docker-compose run --rm certbot certonly --webroot \\"
echo "       --webroot-path=/var/www/certbot \\"
echo "       --email your-email@example.com \\"
echo "       --agree-tos \\"
echo "       --no-eff-email \\"
echo "       -d your-domain.com"
echo "  4. Update nginx.conf to use new certificate"
echo "  5. Restart: docker-compose restart nginx"
echo ""
echo "Note: Domain-based certificates are recommended for production."
echo ""
SSL_EOF

chmod +x setup-ssl.sh

# Create init script to create supertokens database
cat > init-db.sh <<'INIT_EOF'
#!/bin/bash
# Wait for postgres to be ready
until docker exec rex-postgres pg_isready -U %s; do
  echo "Waiting for postgres..."
  sleep 2
done

# Create supertokens database
docker exec rex-postgres psql -U %s -c "CREATE DATABASE %s;" || echo "Database already exists"
echo "Databases initialized"
INIT_EOF

chmod +x init-db.sh

# Create admin initialization script
cat > init-admin.sh <<'ADMIN_EOF'
#!/bin/bash
set -e

echo "=== Initializing Platform Admin User ==="

# Wait for API to be healthy
echo "Waiting for API to be ready..."
max_retries=60
retry_count=0
until curl -f http://localhost:8080/health > /dev/null 2>&1; do
  retry_count=$((retry_count + 1))
  if [ $retry_count -ge $max_retries ]; then
    echo "ERROR: API did not become healthy within timeout"
    exit 1
  fi
  echo "API not ready yet, waiting... ($retry_count/$max_retries)"
  sleep 5
done

echo "API is healthy, proceeding with admin user creation..."

# Wait for SuperTokens to be ready
echo "Waiting for SuperTokens to be ready..."
retry_count=0
until curl -f http://localhost:3567/hello > /dev/null 2>&1; do
  retry_count=$((retry_count + 1))
  if [ $retry_count -ge $max_retries ]; then
    echo "ERROR: SuperTokens did not become healthy within timeout"
    exit 1
  fi
  echo "SuperTokens not ready yet, waiting... ($retry_count/$max_retries)"
  sleep 5
done

echo "SuperTokens is ready!"

# Create admin user via SuperTokens Core API
echo "Creating admin user in SuperTokens..."
SIGNUP_RESPONSE=$(curl -s -X POST http://localhost:3567/recipe/signup \
  -H "Content-Type: application/json" \
  -H "api-key: %s" \
  -d '{
    "email": "admin@platform.local",
    "password": "admin"
  }')

echo "SuperTokens signup response: $SIGNUP_RESPONSE"

# Extract user ID from response
USER_ID=$(echo "$SIGNUP_RESPONSE" | jq -r '.user.id // empty')

if [ -z "$USER_ID" ]; then
  # Check if user already exists
  if echo "$SIGNUP_RESPONSE" | grep -q "EMAIL_ALREADY_EXISTS_ERROR"; then
    echo "Admin user already exists in SuperTokens, fetching user ID..."
    
    # Sign in to get user ID
    SIGNIN_RESPONSE=$(curl -s -X POST http://localhost:3567/recipe/signin \
      -H "Content-Type: application/json" \
      -H "api-key: %s" \
      -d '{
        "email": "admin@platform.local",
        "password": "admin"
      }')
    
    USER_ID=$(echo "$SIGNIN_RESPONSE" | jq -r '.user.id // empty')
    
    if [ -z "$USER_ID" ]; then
      echo "ERROR: Failed to get user ID from existing user"
      echo "Sign in response: $SIGNIN_RESPONSE"
      exit 1
    fi
    
    echo "Found existing user ID: $USER_ID"
  else
    echo "ERROR: Failed to create admin user"
    echo "Response: $SIGNUP_RESPONSE"
    exit 1
  fi
else
  echo "Successfully created admin user with ID: $USER_ID"
fi

# Set user metadata to mark as platform admin
echo "Setting user metadata..."
curl -s -X PUT "http://localhost:3567/recipe/user/metadata" \
  -H "Content-Type: application/json" \
  -H "api-key: %s" \
  -d "{
    \"userId\": \"$USER_ID\",
    \"metadataUpdate\": {
      \"is_platform_admin\": true,
      \"created_by\": \"system\",
      \"email\": \"admin@platform.local\"
    }
  }" > /dev/null

echo "User metadata set successfully"

# Add user to platform_admins table in database
echo "Adding user to platform_admins table..."
docker exec rex-postgres psql -U %s -d %s -c "
  INSERT INTO platform_admins (user_id, created_by) 
  VALUES ('$USER_ID', 'system') 
  ON CONFLICT (user_id) DO NOTHING;
" > /dev/null 2>&1

if [ $? -eq 0 ]; then
  echo "✓ Successfully added user to platform_admins table"
else
  echo "⚠ Warning: Failed to add user to platform_admins table (may already exist)"
fi

echo "=== Platform Admin User Initialization Complete ==="
echo ""
echo "Admin credentials:"
echo "  Email: admin@platform.local"
echo "  Password: admin"
echo "  User ID: $USER_ID"
echo ""
echo "⚠ IMPORTANT: Please change the default password immediately after first login!"
echo ""

# Create a status file to indicate initialization is complete
touch /app/.admin-initialized
ADMIN_EOF

chmod +x init-admin.sh

# Pull images
docker-compose pull postgres redis supertokens
docker pull %s:latest
docker pull %s:latest

# Start services
docker-compose up -d postgres redis

# Initialize databases
sleep 10
./init-db.sh

# Start remaining services
docker-compose up -d

# Setup SSL certificates
echo "Setting up SSL..."
./setup-ssl.sh

# Wait a bit for services to stabilize, then initialize admin user
sleep 30
echo "Running admin initialization..."
./init-admin.sh || echo "Admin initialization failed, check logs at /app/admin-init.log"

# Setup log rotation
cat > /etc/logrotate.d/docker-compose <<'LOGROTATE_EOF'
/var/lib/docker/containers/*/*.log {
    rotate 7
    daily
    compress
    size=10M
    missingok
    delaycompress
    copytruncate
}
LOGROTATE_EOF

# Create update script
cat > /app/update.sh <<'UPDATE_EOF'
#!/bin/bash
# Update script to pull latest images and restart
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin $(echo %s | cut -d'/' -f1)
docker-compose pull api worker
docker-compose up -d api worker
docker image prune -f
UPDATE_EOF

chmod +x /app/update.sh

# Install CloudWatch agent for monitoring (optional)
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
dpkg -i -E ./amazon-cloudwatch-agent.deb

echo "All-in-one setup complete!"
`
			script := fmt.Sprintf(scriptTemplate, apiRepo, username, password, mainDB, username, username, password, supertokensDB, apiKey,
				apiRepo, username, password, mainDB, apiKey, workerRepo, username, password, mainDB,
				username, username, supertokensDB, apiKey, apiKey, apiKey, username, mainDB, apiRepo, workerRepo, apiRepo)

			// Compress the script using gzip to stay under 16KB limit
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			if _, err := gz.Write([]byte(script)); err != nil {
				return "", err
			}
			if err := gz.Close(); err != nil {
				return "", err
			}

			// Return base64-encoded gzipped script with proper MIME type
			compressed := base64.StdEncoding.EncodeToString(buf.Bytes())

			// Use cloud-init MIME format for gzipped data
			return fmt.Sprintf(`Content-Type: multipart/mixed; boundary="===============BOUNDARY=="
MIME-Version: 1.0

--===============BOUNDARY==
Content-Type: text/x-shellscript; charset="us-ascii"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="userdata.sh"

%s
--===============BOUNDARY==--
`, compressed), nil
		}).(pulumi.StringOutput)

	// Create EC2 Instance (On-Demand for easier updates, or Spot for lower cost)
	// Note: Spot instances cannot be stopped/started for user data updates
	// Use on-demand if you need to update frequently, or replace spot instances
	useSpot := true // Set to false for on-demand instance

	instanceArgs := &ec2.InstanceArgs{
		Ami:                pulumi.String(ami.Id),
		InstanceType:       pulumi.String("t4g.medium"), // 2 vCPU, 4 GB RAM (ARM Graviton)
		IamInstanceProfile: instanceProfile.Name,
		SubnetId: network.PublicSubnetIDs.ApplyT(func(subnets []string) string {
			return subnets[0]
		}).(pulumi.StringOutput),
		VpcSecurityGroupIds:      pulumi.StringArray{ec2SG.ID().ToStringOutput()},
		AssociatePublicIpAddress: pulumi.Bool(true), // Assign public IP for direct access
		UserData:                 userData,

		// Root volume configuration
		RootBlockDevice: &ec2.InstanceRootBlockDeviceArgs{
			VolumeSize:          pulumi.Int(30),
			VolumeType:          pulumi.String("gp3"),
			DeleteOnTermination: pulumi.Bool(true),
		},

		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-allinone", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
			"Type":        pulumi.String("allinone-docker-compose"),
		},
	}

	// Add spot configuration if enabled
	if useSpot {
		instanceArgs.InstanceMarketOptions = &ec2.InstanceInstanceMarketOptionsArgs{
			MarketType: pulumi.String("spot"),
			SpotOptions: &ec2.InstanceInstanceMarketOptionsSpotOptionsArgs{
				MaxPrice:                     pulumi.String("0.05"),
				SpotInstanceType:             pulumi.String("persistent"),
				InstanceInterruptionBehavior: pulumi.String("stop"),
			},
		}
	}

	instance, err := ec2.NewInstance(ctx, fmt.Sprintf("%s-%s-allinone-instance", projectName, environment), instanceArgs)
	if err != nil {
		return nil, err
	}

	// Create Elastic IP for persistent public IP across instance replacements
	eip, err := ec2.NewEip(ctx, fmt.Sprintf("%s-%s-allinone-eip", projectName, environment), &ec2.EipArgs{
		Instance: instance.ID(),
		Domain:   pulumi.String("vpc"),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-allinone-eip", projectName, environment)),
			"Project":     tags["Project"],
			"Environment": tags["Environment"],
			"ManagedBy":   tags["ManagedBy"],
		},
	})
	if err != nil {
		return nil, err
	}

	return &AllInOneResources{
		Instance:          instance,
		SecurityGroup:     ec2SG,
		PublicIP:          eip.PublicIp,
		PrivateIP:         instance.PrivateIp,
		PublicDNS:         eip.PublicDns,
		MasterUsername:    masterUsername,
		MainDBName:        mainDBName,
		SuperTokensDBName: supertokensDBName,
		MasterPassword:    masterPassword,
	}, nil
}
