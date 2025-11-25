# Database Migrations

Complete guide to managing database migrations.

## Overview

This project uses **golang-migrate** for database schema management. Migrations are:
- Versioned SQL files
- Applied sequentially
- Reversible (up and down)
- Tracked in `schema_migrations` table

## Migration Tools

### CLI Tool

Located at: `cmd/migrate/main.go`

```bash
# Run migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status

# Create new migration
make migrate-create name=add_users_table
```

### Migration Files

Located in: `migrations/`

```
migrations/
├── 20241101120000_create_tenants.up.sql
├── 20241101120000_create_tenants.down.sql
├── 20241102130000_create_tenant_members.up.sql
├── 20241102130000_create_tenant_members.down.sql
└── ...
```

**Naming Convention**: `YYYYMMDDHHMMSS_description.up.sql` / `.down.sql`

## Creating Migrations

### Generate Migration Files

```bash
make migrate-create name=add_articles_table
```

This creates:
- `migrations/YYYYMMDDHHMMSS_add_articles_table.up.sql`
- `migrations/YYYYMMDDHHMMSS_add_articles_table.down.sql`

### Write Up Migration

```sql
-- migrations/20241125100000_add_articles_table.up.sql

-- Create articles table
CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    author_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_articles_tenant_id ON articles(tenant_id);
CREATE INDEX idx_articles_author_id ON articles(author_id);
CREATE INDEX idx_articles_status ON articles(status);

-- Add updated_at trigger
CREATE TRIGGER update_articles_updated_at
    BEFORE UPDATE ON articles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Write Down Migration

```sql
-- migrations/20241125100000_add_articles_table.down.sql

-- Drop trigger
DROP TRIGGER IF EXISTS update_articles_updated_at ON articles;

-- Drop indexes
DROP INDEX IF EXISTS idx_articles_status;
DROP INDEX IF EXISTS idx_articles_author_id;
DROP INDEX IF EXISTS idx_articles_tenant_id;

-- Drop table
DROP TABLE IF EXISTS articles;
```

## Running Migrations

### Development

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Rollback all migrations
make migrate-down-all
```

### Production (Docker)

```bash
# SSH into EC2/ECS
ssh ec2-user@<ip>

# Run migrations in container
docker exec utm-api ./migrate up

# Or via make
docker exec utm-api make migrate-up
```

### Production (ECS Task)

```bash
# Run migration task
aws ecs run-task \
  --cluster utm-cluster \
  --task-definition utm-migration \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx]}"
```

## Migration Status

### Check Current Version

```bash
make migrate-status
```

Output:
```
Current version: 20241125100000
Pending migrations: 0
```

### View Migration History

```sql
-- Connect to database
psql -h localhost -U postgres -d utm_backend

-- Check schema_migrations table
SELECT * FROM schema_migrations ORDER BY version DESC;
```

Output:
```
 version        | dirty
----------------+-------
 20241125100000 | f
 20241124150000 | f
 20241123120000 | f
```

## Common Migrations

### Add Column

**Up**:
```sql
ALTER TABLE tenants ADD COLUMN description TEXT;
```

**Down**:
```sql
ALTER TABLE tenants DROP COLUMN description;
```

### Modify Column

**Up**:
```sql
ALTER TABLE tenants ALTER COLUMN name TYPE VARCHAR(255);
ALTER TABLE tenants ALTER COLUMN name SET NOT NULL;
```

**Down**:
```sql
ALTER TABLE tenants ALTER COLUMN name DROP NOT NULL;
ALTER TABLE tenants ALTER COLUMN name TYPE VARCHAR(100);
```

### Add Index

**Up**:
```sql
CREATE INDEX CONCURRENTLY idx_tenants_slug ON tenants(slug);
```

**Down**:
```sql
DROP INDEX CONCURRENTLY IF EXISTS idx_tenants_slug;
```

### Add Foreign Key

**Up**:
```sql
ALTER TABLE articles
ADD CONSTRAINT fk_articles_tenant
FOREIGN KEY (tenant_id)
REFERENCES tenants(id)
ON DELETE CASCADE;
```

**Down**:
```sql
ALTER TABLE articles DROP CONSTRAINT IF EXISTS fk_articles_tenant;
```

### Create Enum Type

**Up**:
```sql
CREATE TYPE article_status AS ENUM ('draft', 'published', 'archived');
ALTER TABLE articles ALTER COLUMN status TYPE article_status USING status::article_status;
```

**Down**:
```sql
ALTER TABLE articles ALTER COLUMN status TYPE VARCHAR(20);
DROP TYPE IF EXISTS article_status;
```

## Data Migrations

### Backfill Data

**Up**:
```sql
-- Add column
ALTER TABLE tenants ADD COLUMN plan VARCHAR(50);

-- Backfill existing data
UPDATE tenants SET plan = 'free' WHERE plan IS NULL;

-- Make NOT NULL
ALTER TABLE tenants ALTER COLUMN plan SET NOT NULL;
```

**Down**:
```sql
ALTER TABLE tenants DROP COLUMN plan;
```

### Transform Data

**Up**:
```sql
-- Create new column
ALTER TABLE users ADD COLUMN full_name VARCHAR(255);

-- Migrate data
UPDATE users SET full_name = first_name || ' ' || last_name;

-- Drop old columns
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
```

**Down**:
```sql
-- Recreate columns
ALTER TABLE users ADD COLUMN first_name VARCHAR(100);
ALTER TABLE users ADD COLUMN last_name VARCHAR(100);

-- Reverse migration
UPDATE users SET
  first_name = SPLIT_PART(full_name, ' ', 1),
  last_name = SPLIT_PART(full_name, ' ', 2);

-- Drop combined column
ALTER TABLE users DROP COLUMN full_name;
```

## Best Practices

### 1. Always Write Down Migrations

```sql
-- ✅ Good: Reversible
-- UP: Add column
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- DOWN: Remove column
ALTER TABLE users DROP COLUMN phone;

-- ❌ Bad: No down migration
-- Only UP migration provided
```

### 2. Use Transactions

```sql
-- Wrap in transaction
BEGIN;

ALTER TABLE users ADD COLUMN status VARCHAR(20);
UPDATE users SET status = 'active';
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

COMMIT;
```

### 3. Test Rollback

```bash
# Apply migration
make migrate-up

# Test rollback
make migrate-down

# Reapply
make migrate-up
```

### 4. Keep Migrations Small

```
✅ Good: One logical change per migration
- 001_create_users.sql
- 002_add_users_email_index.sql
- 003_add_users_status.sql

❌ Bad: Multiple unrelated changes
- 001_big_migration.sql (creates 10 tables, adds indexes, migrates data)
```

### 5. Never Modify Applied Migrations

```
❌ Don't edit: 20241101_create_users.up.sql (already applied)
✅ Create new: 20241125_modify_users.up.sql
```

## Troubleshooting

### Dirty Migration

**Symptom**: Migration fails, `dirty = true` in schema_migrations

**Cause**: Migration partially applied and failed

**Solution**:
```sql
-- Check which migration is dirty
SELECT * FROM schema_migrations WHERE dirty = true;

-- Manually fix the issue, then reset dirty flag
UPDATE schema_migrations SET dirty = false WHERE version = '<version>';

-- Or rollback to previous version
make migrate-down
```

### Migration Already Applied

**Symptom**: "migration already applied" error

**Cause**: Migration file renamed or duplicated

**Solution**:
```bash
# Check schema_migrations table
psql -h localhost -U postgres -d utm_backend -c "SELECT * FROM schema_migrations;"

# Remove duplicate entry if needed
psql -h localhost -U postgres -d utm_backend -c "DELETE FROM schema_migrations WHERE version = '<version>';"
```

### Foreign Key Constraint Error

**Symptom**: Cannot drop table due to foreign keys

**Solution**:
```sql
-- Drop dependent foreign keys first
ALTER TABLE articles DROP CONSTRAINT fk_articles_tenant;

-- Then drop table
DROP TABLE tenants;
```

### Slow Migration

**Symptom**: Migration takes very long on large tables

**Solution**:
```sql
-- Use CONCURRENTLY for indexes (doesn't lock table)
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- For large data migrations, batch updates
DO $$
DECLARE
  batch_size INT := 1000;
  offset_val INT := 0;
BEGIN
  LOOP
    UPDATE users SET status = 'active'
    WHERE id IN (
      SELECT id FROM users WHERE status IS NULL
      LIMIT batch_size OFFSET offset_val
    );
    
    IF NOT FOUND THEN EXIT; END IF;
    offset_val := offset_val + batch_size;
  END LOOP;
END $$;
```

## Production Checklist

Before running migrations in production:

- [ ] Test migrations in staging environment
- [ ] Test rollback (down migration)
- [ ] Backup database
- [ ] Check migration doesn't lock critical tables
- [ ] Schedule during low-traffic period
- [ ] Monitor during migration
- [ ] Verify data integrity after migration
- [ ] Test application functionality

## Advanced Patterns

### Zero-Downtime Migrations

**Add column (nullable first)**:
```sql
-- Step 1: Add nullable column
ALTER TABLE users ADD COLUMN email_verified BOOLEAN;

-- Deploy application version that writes to new column

-- Step 2: Backfill existing data
UPDATE users SET email_verified = false WHERE email_verified IS NULL;

-- Step 3: Make NOT NULL
ALTER TABLE users ALTER COLUMN email_verified SET NOT NULL;
```

### Rename Column Without Downtime

```sql
-- Step 1: Add new column
ALTER TABLE users ADD COLUMN username VARCHAR(100);

-- Step 2: Copy data
UPDATE users SET username = user_name;

-- Deploy application version that uses both columns

-- Step 3: Drop old column (after verification)
ALTER TABLE users DROP COLUMN user_name;
```

## Next Steps

- [Environment Configuration](/deployment/environment) - Database configuration
- [Production Setup](/deployment/production-setup) - Deployment guide
- [Docker Deployment](/deployment/docker) - Running migrations in Docker
- [AWS Deployment](/deployment/aws) - ECS migration tasks
