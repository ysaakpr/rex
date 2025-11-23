# Networking Architecture: All-in-One vs Standard Mode

## Overview

This document explains the networking differences between deployment modes and clarifies why NAT Gateway is not needed for all-in-one deployments.

## Deployment Modes Comparison

### 1. Standard/Low-Cost Mode (Fargate)

**Architecture:**
```
Internet
    ↓
Application Load Balancer (ALB)
    ├─ Port 80/443 → Routes to Fargate tasks
    └─ Health checks
    ↓
Private Subnets (10.0.10.0/24, 10.0.11.0/24)
    ├─ Fargate Task: API
    ├─ Fargate Task: Worker
    ├─ Fargate Task: SuperTokens
    └─ (Low-cost: EC2 with PostgreSQL/Redis)
    ↓ (outbound internet)
NAT Gateway (~$32/mo)
    ↓
Internet Gateway (FREE)
    ↓
Internet
```

**Network Flow:**
- **Inbound**: Internet → ALB → Private Subnet → Fargate Tasks
- **Outbound**: Fargate Tasks → NAT Gateway → Internet Gateway → Internet

**Why This Setup?**
- Fargate tasks MUST be in private subnets for security
- ALB provides SSL termination and path-based routing
- NAT Gateway enables outbound internet (apt updates, ECR pulls, etc.)
- More secure but more expensive

**Monthly Costs:**
- ALB: ~$16/month ($0.0225/hour)
- NAT Gateway: ~$32/month ($0.045/hour + $0.045/GB)
- Data processing: Variable
- **Total Networking: ~$48-60/month**

### 2. All-in-One Mode (Docker Compose)

**Architecture:**
```
Internet
    ↓
Internet Gateway (FREE - included with VPC)
    ↓
Public Subnet (10.0.0.0/24)
    ↓
EC2 Instance (Public IP: a.b.c.d)
    ├─ Security Group (port-based filtering)
    ├─ Port 8080 → API container
    ├─ Port 3567 → SuperTokens container
    ├─ Port 80 → Frontend (optional)
    └─ Docker Compose internal networking
    ↓ (outbound internet - same path)
Internet Gateway (FREE)
    ↓
Internet
```

**Network Flow:**
- **Inbound**: Internet → Internet Gateway → Public Subnet → EC2 Instance:Port
- **Outbound**: EC2 Instance → Internet Gateway → Internet

**Why This Setup?**
- Single EC2 instance = can be in public subnet
- Direct port access = no ALB needed
- Public IP = uses Internet Gateway (free)
- Security groups control access (like a firewall)
- Simpler and much cheaper

**Monthly Costs:**
- Internet Gateway: FREE (included with VPC)
- Public IP: FREE (included with instance)
- Security Group: FREE
- **Total Networking: $0/month**

## NAT Gateway Deep Dive

### What is NAT Gateway?

**Network Address Translation (NAT) Gateway** allows instances in **private subnets** to initiate outbound connections to the internet while preventing unsolicited inbound connections.

### When DO You Need NAT Gateway?

✅ **Required when:**
- Instances are in **private subnets**
- Need outbound internet access (updates, API calls, downloads)
- Don't want direct internet exposure
- Running Fargate tasks (must be in private subnets)
- Running RDS, ElastiCache in private subnets (they need updates)

### When DON'T You Need NAT Gateway?

❌ **Not needed when:**
- Instances are in **public subnets** with **public IPs**
- Using **Internet Gateway** directly (free)
- All services on one EC2 instance (all-in-one mode)
- Don't need to hide instance IPs

### NAT Gateway Costs

**Pricing (us-east-1 region):**
- Hourly charge: $0.045/hour = **$32.40/month**
- Data processing: $0.045/GB processed
- Always running (no auto-scaling)

**Example monthly cost:**
- Base: $32.40
- 100 GB data: $4.50
- **Total: ~$37/month** for moderate usage

## Internet Gateway vs NAT Gateway

| Feature | Internet Gateway | NAT Gateway |
|---------|-----------------|-------------|
| **Cost** | FREE | $32/month + data |
| **Purpose** | Direct internet access | Masked outbound access |
| **Direction** | Bidirectional | Outbound only |
| **IP** | Uses public IPs | Provides translation |
| **Use Case** | Public subnets | Private subnets |
| **Management** | Fully managed (AWS) | Fully managed (AWS) |
| **Availability** | Region-level | AZ-specific |

## Security Considerations

### Public Subnet Security (All-in-One)

**Concerns:**
- ⚠️ Direct internet exposure
- ⚠️ All ports potentially accessible
- ⚠️ Instance IP is public

**Mitigations:**
- ✅ Security Group rules (like firewall)
- ✅ Only open required ports (8080, 3567, 80, 22)
- ✅ SSH via SSM (no port 22 needed)
- ✅ Can restrict by IP ranges if needed
- ✅ Regular security updates
- ✅ Fail2ban for brute force protection

**Security Group Example:**
```hcl
Ingress Rules:
- Port 8080 (API): 0.0.0.0/0 (or restrict to specific IPs)
- Port 3567 (SuperTokens): 0.0.0.0/0 (or restrict to frontend IPs)
- Port 80 (HTTP): 0.0.0.0/0
- Port 443 (HTTPS): 0.0.0.0/0
- Port 22 (SSH): Can be removed if using SSM

Egress Rules:
- All traffic: 0.0.0.0/0 (for updates, Docker pulls, etc.)
```

### Private Subnet Security (Standard Mode)

**Benefits:**
- ✅ No direct internet exposure
- ✅ ALB provides additional security layer
- ✅ Can use WAF (Web Application Firewall)
- ✅ SSL termination at ALB
- ✅ DDoS protection via Shield

**Trade-offs:**
- 💰 More expensive (~$48+/month for networking)
- 🔧 More complex setup
- 🔧 Harder to debug (need bastion host or SSM)

## Cost Analysis

### Scenario: Small Development Environment

**Standard Mode (Fargate + ALB + NAT):**
```
ALB:                $16/month
NAT Gateway:        $32/month
Data processing:    $5/month (estimated)
Fargate tasks:      $45/month (3 tasks)
Total:              $98/month
```

**All-in-One Mode (EC2 + Public IP):**
```
EC2 t3a.medium:     $10/month (spot)
Internet Gateway:   FREE
Public IP:          FREE
EBS storage:        $2.40/month
Total:              $12.40/month
```

**Savings: $85.60/month (87% reduction!)**

### Scenario: Production Environment

**Standard Mode (Recommended):**
- Multiple AZs for high availability
- ALB with SSL certificate
- NAT Gateway in each AZ (high availability)
- RDS Aurora Serverless
- ElastiCache Redis
- **Cost: $200-500/month** depending on traffic

**All-in-One Mode (NOT Recommended for Production):**
- Single point of failure
- No automatic failover
- Limited by single instance resources
- But **only $12/month** if you accept the risks

## Best Practices

### Use All-in-One Mode When:

✅ Development environments
✅ Testing and staging
✅ Personal projects
✅ Low-traffic applications
✅ Cost is primary concern
✅ Downtime is acceptable
✅ Single region deployment

### Use Standard Mode When:

✅ Production workloads
✅ High availability required
✅ Compliance requirements (private subnets)
✅ Need WAF or advanced security
✅ Multi-AZ deployment
✅ Auto-scaling needed
✅ 24/7 uptime critical

## Migration Path

### From Standard to All-in-One

1. **Backup everything** (database, configs)
2. Set `pulumi config set rex-backend:allinone true`
3. Run `pulumi up`
4. Update DNS to point to instance public IP:8080
5. Test thoroughly
6. Destroy old resources: `pulumi destroy` (with allinone=false)

### From All-in-One to Standard

1. Set `pulumi config set rex-backend:allinone false`
2. Configure database credentials
3. Run `pulumi up` (creates ALB, NAT, Fargate)
4. Migrate database data
5. Update DNS to point to ALB
6. Test thoroughly
7. Destroy old EC2 instance

## Troubleshooting

### All-in-One Mode Issues

**Cannot connect to API:**
```bash
# Check instance has public IP
pulumi stack output allInOnePublicIp

# Check security group allows port 8080
aws ec2 describe-security-groups --filters "Name=tag:Name,Values=*allinone-sg*"

# Test from different location
curl -v http://PUBLIC_IP:8080/health

# Check if services are running
aws ssm start-session --target INSTANCE_ID
docker-compose ps
```

**High data transfer costs:**
- Check CloudWatch metrics for data out
- Consider restricting access by IP
- Use CloudFront if serving large files
- Monitor security group logs

### Standard Mode Issues

**High NAT Gateway costs:**
- Check data processing charges
- Reduce unnecessary outbound traffic
- Use VPC endpoints for AWS services (S3, DynamoDB, etc.)
- Consider caching to reduce external API calls

**ALB costs too high:**
- Consolidate to single ALB if possible
- Use path-based routing
- Consider Application Load Balancer vs Classic Load Balancer
- Check LCU (Load Balancer Capacity Units) usage

## Conclusion

The all-in-one deployment mode eliminates the need for both **ALB** and **NAT Gateway** by:
1. Placing the EC2 instance in a **public subnet**
2. Assigning a **public IP address**
3. Using **Internet Gateway** for all traffic (free)
4. Implementing **security groups** for access control

This results in **~$48-60/month savings** on networking costs alone, making it perfect for development, testing, and low-traffic production workloads where the cost savings outweigh the high-availability benefits of the standard architecture.

For production workloads requiring high availability, auto-scaling, and enhanced security, the standard mode with ALB and NAT Gateway is recommended despite the higher costs.

---

**Last Updated**: November 23, 2025  
**Related Docs**:
- [ALLINONE_QUICKSTART.md](./ALLINONE_QUICKSTART.md)
- [LOWCOST_ALLINONE.md](./LOWCOST_ALLINONE.md)
- [AWS VPC Documentation](https://docs.aws.amazon.com/vpc/)
- [AWS NAT Gateway Pricing](https://aws.amazon.com/vpc/pricing/)

