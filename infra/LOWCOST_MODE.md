# Low-Cost Deployment Mode

## Overview

The infrastructure now supports a **low-cost deployment mode** that significantly reduces AWS costs by replacing managed services (RDS Aurora and ElastiCache) with a self-hosted PostgreSQL and Redis instance running on a single EC2 Spot instance.

## Cost Comparison

### Standard Mode
- **RDS Aurora Serverless v2**: ~$50-100/month (minimum 0.5 ACU)
- **ElastiCache Redis**: ~$15-30/month (cache.t4g.micro)
- **Total**: ~$65-130/month for database services

### Low-Cost Mode
- **EC2 Spot Instance (t3a.small)**: ~$3-8/month (spot pricing with interruptions)
- **EBS Storage (30GB gp3)**: ~$2.40/month
- **Total**: ~$5-10/month for database services

**Savings**: ~85-90% cost reduction on database infrastructure

## Architecture

### Low-Cost Mode
```
┌─────────────────────────────────┐
│   EC2 Spot Instance (t3a.small) │
│   - 2 vCPU, 2 GB RAM            │
│   - Ubuntu 22.04 LTS            │
│   ├─ PostgreSQL 14              │
│   │  ├─ rex_backend DB          │
│   │  └─ supertokens DB          │
│   └─ Redis Server               │
│      └─ 512 MB max memory       │
└─────────────────────────────────┘
```

### Instance Details
- **Instance Type**: `t3a.small` (2 vCPU, 2 GB RAM)
- **Pricing**: Spot instance (up to 90% cheaper than on-demand)
- **Interruption Behavior**: Stop (not terminate) - data is preserved
- **Max Price**: $0.05/hour
- **Storage**: 30 GB GP3 EBS volume
- **Network**: Deployed in private subnet

## Configuration

### Enable Low-Cost Mode

Add the following to your Pulumi configuration:

```bash
# Set low-cost mode to true
pulumi config set rex-backend:lowcost true
```

Or update your `Pulumi.dev.yaml`:

```yaml
config:
  rex-backend:lowcost: "true"
```

### Disable Low-Cost Mode (Use Managed Services)

```bash
# Set low-cost mode to false or remove the config
pulumi config set rex-backend:lowcost false
```

## Deployment

### Initial Deployment

```bash
# Navigate to infrastructure directory
cd infra

# Configure low-cost mode
pulumi config set rex-backend:lowcost true

# Set required secrets (if not already set)
pulumi config set --secret rex-backend:dbMasterPassword "your-secure-password"
pulumi config set --secret rex-backend:supertokensApiKey "your-api-key"

# Deploy
pulumi up
```

### Check Deployment Status

After deployment, check the exported values:

```bash
pulumi stack output
```

You should see:
- `deploymentMode`: "lowcost"
- `lowCostInstanceId`: EC2 instance ID
- `lowCostInstancePrivateIp`: Private IP address
- `databaseEndpoint`: Same as lowCostInstancePrivateIp
- `redisEndpoint`: Same as lowCostInstancePrivateIp

## Features

### Automatic Setup

The EC2 instance is automatically configured with:

1. **PostgreSQL 14**
   - Listening on all interfaces (0.0.0.0)
   - MD5 password authentication
   - Two databases: `rex_backend` and `supertokens`
   - User with superuser privileges

2. **Redis Server**
   - Listening on all interfaces (0.0.0.0)
   - Protected mode disabled (private network only)
   - Max memory: 512 MB
   - Eviction policy: allkeys-lru

3. **CloudWatch Agent**
   - Installed for monitoring (optional)

4. **Systems Manager (SSM)**
   - Enabled for secure shell access
   - No SSH key required

### Security

- **Network Isolation**: Instance deployed in private subnet
- **Security Group**: Only allows PostgreSQL (5432) and Redis (6379) from ECS tasks
- **SSH Access**: Disabled by default (use SSM Session Manager instead)
- **No Public IP**: Instance is not directly accessible from internet

### Access the Instance

Use AWS Systems Manager Session Manager (no SSH key needed):

```bash
# Get instance ID
INSTANCE_ID=$(pulumi stack output lowCostInstanceId)

# Connect via SSM
aws ssm start-session --target $INSTANCE_ID

# Once connected, you can:
# - Check PostgreSQL: sudo -u postgres psql
# - Check Redis: redis-cli ping
# - View logs: journalctl -u postgresql -u redis-server
```

## Limitations & Considerations

### ⚠️ Important Limitations

1. **Spot Instance Interruptions**
   - Spot instances can be interrupted with 2-minute notice
   - Instance will **stop** (not terminate) - data is preserved
   - AWS will attempt to restart when capacity is available
   - Expect occasional downtime (typically rare for t3a.small)

2. **Single Point of Failure**
   - No high availability or automatic failover
   - No read replicas
   - Suitable for dev/test/staging environments only

3. **Performance**
   - Limited to 2 vCPU and 2 GB RAM
   - Suitable for light to moderate workloads
   - No auto-scaling capabilities

4. **Backup & Recovery**
   - No automated backups (unlike RDS)
   - You must implement your own backup strategy
   - EBS snapshots recommended

5. **SSL/TLS**
   - PostgreSQL configured without SSL by default
   - Connections use MD5 password auth
   - Fine for private network, but less secure than RDS

### ✅ Best Use Cases

- **Development environments**: Perfect for dev/test
- **Staging environments**: Good for pre-production testing
- **Small projects**: Suitable for MVPs and small applications
- **Learning/experimentation**: Great for cost-conscious learning

### ❌ Not Recommended For

- **Production environments**: Use managed services (RDS + ElastiCache)
- **Mission-critical applications**: High availability is essential
- **Compliance requirements**: Managed services offer better compliance
- **High-traffic applications**: Need better scaling and performance

## Backup Strategy

Since RDS automated backups are not available, implement manual backups:

### 1. EBS Snapshots

```bash
# Get volume ID
INSTANCE_ID=$(pulumi stack output lowCostInstanceId)
VOLUME_ID=$(aws ec2 describe-instances --instance-ids $INSTANCE_ID \
  --query 'Reservations[0].Instances[0].BlockDeviceMappings[0].Ebs.VolumeId' \
  --output text)

# Create snapshot
aws ec2 create-snapshot --volume-id $VOLUME_ID \
  --description "rex-backend-lowcost-$(date +%Y%m%d-%H%M%S)"
```

### 2. PostgreSQL Dumps

```bash
# Connect to instance
aws ssm start-session --target $(pulumi stack output lowCostInstanceId)

# Create backup (on the instance)
sudo -u postgres pg_dump rex_backend > /tmp/rex_backend_backup.sql
sudo -u postgres pg_dump supertokens > /tmp/supertokens_backup.sql

# Copy to S3 (if configured)
aws s3 cp /tmp/rex_backend_backup.sql s3://your-backup-bucket/
aws s3 cp /tmp/supertokens_backup.sql s3://your-backup-bucket/
```

### 3. Automated Backup Script

Consider creating a Lambda function or cron job for automated backups.

## Monitoring

### CloudWatch Metrics

The instance exports standard EC2 metrics:
- CPU utilization
- Network traffic
- Disk I/O

### Custom Monitoring

Set up CloudWatch alarms for:
- High CPU usage (>80%)
- Low disk space (<10% free)
- Instance status checks

### Application Health

Monitor your application logs to detect:
- Database connection failures (spot interruption)
- Slow queries
- Redis connection issues

## Migration Between Modes

### From Standard to Low-Cost

**⚠️ Warning**: This will destroy your RDS and ElastiCache data!

```bash
# 1. Backup your data first!
# 2. Update configuration
pulumi config set rex-backend:lowcost true

# 3. Deploy (this will destroy RDS and ElastiCache)
pulumi up

# 4. Restore data to new EC2 instance
```

### From Low-Cost to Standard

```bash
# 1. Backup EC2 data first!
# 2. Update configuration
pulumi config set rex-backend:lowcost false

# 3. Deploy (this will destroy EC2 instance)
pulumi up

# 4. Restore data to RDS
```

## Troubleshooting

### Instance Not Starting

```bash
# Check spot request status
aws ec2 describe-spot-instance-requests

# Check instance status
aws ec2 describe-instance-status --instance-ids $(pulumi stack output lowCostInstanceId)
```

### Database Connection Issues

```bash
# Connect to instance
aws ssm start-session --target $(pulumi stack output lowCostInstanceId)

# Check PostgreSQL status
sudo systemctl status postgresql

# Check PostgreSQL logs
sudo journalctl -u postgresql -n 50

# Test connection
psql -h localhost -U rexadmin -d rex_backend
```

### Redis Connection Issues

```bash
# Check Redis status
sudo systemctl status redis-server

# Check Redis logs
sudo journalctl -u redis-server -n 50

# Test connection
redis-cli ping
```

### Spot Instance Interrupted

If your spot instance is interrupted:

1. **Check spot interruption notice**:
   ```bash
   aws ec2 describe-instance-status --instance-ids $INSTANCE_ID
   ```

2. **Wait for instance to restart** (usually automatic)

3. **Manually start if needed**:
   ```bash
   aws ec2 start-instances --instance-ids $INSTANCE_ID
   ```

4. **Consider switching instance type** if interruptions are frequent:
   - Edit `infra/ec2_lowcost.go`
   - Change `t3a.small` to `t3.small` or `t3a.micro`

## Cost Optimization Tips

1. **Use t3a instead of t3** (AMD instances are ~10% cheaper)
2. **Enable EBS snapshot lifecycle policy** (delete old backups)
3. **Stop instance during non-work hours** (dev environments)
4. **Monitor spot pricing trends** in your region
5. **Consider Reserved Instances** if you need 24/7 uptime

## Production Readiness Checklist

If you want to use low-cost mode in production (not recommended):

- [ ] Implement automated backup strategy
- [ ] Set up monitoring and alerting
- [ ] Test spot interruption recovery
- [ ] Document recovery procedures
- [ ] Configure CloudWatch alarms
- [ ] Enable EBS encryption
- [ ] Set up PostgreSQL SSL
- [ ] Implement Redis authentication
- [ ] Test performance under load
- [ ] Plan for maintenance windows
- [ ] Have fallback plan to migrate to managed services

## Support

For issues or questions:
1. Check CloudWatch logs
2. Use SSM Session Manager to debug
3. Review Pulumi logs: `pulumi logs`
4. Check AWS Systems Manager for instance status

## References

- [AWS Spot Instances](https://aws.amazon.com/ec2/spot/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/documentation)
- [AWS Systems Manager Session Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager.html)

