# All-in-One Mode: Removed ALB Dependency

## Change Summary

Removed Application Load Balancer (ALB) requirement from all-in-one deployment mode to reduce costs and simplify architecture. Services are now accessed directly via ports on the EC2 instance's public IP.

## Date
November 23, 2025

## Motivation

**Problem:**
- ALB adds ~$16/month fixed cost
- NAT Gateway adds ~$32/month cost
- All-in-one already consolidates everything on one instance
- ALB routing adds unnecessary complexity for single-instance deployments

**Solution:**
- Move EC2 instance to public subnet
- Assign public IP address
- Expose services on different ports
- Remove ALB entirely
- Use Internet Gateway (free) instead of NAT Gateway

## Changes Made

### 1. Infrastructure Code (`infra/ec2_allinone.go`)

**Subnet Change:**
```go
// Before: Private subnet
SubnetId: network.PrivateSubnetIDs.ApplyT(func(subnets []string) string {
    return subnets[0]
}).(pulumi.StringOutput),

// After: Public subnet with public IP
SubnetId: network.PublicSubnetIDs.ApplyT(func(subnets []string) string {
    return subnets[0]
}).(pulumi.StringOutput),
AssociatePublicIpAddress: pulumi.Bool(true),
```

**Security Group Updates:**
```go
// Removed: ALB-specific rules
SecurityGroups: pulumi.StringArray{albSG.ID().ToStringOutput()}

// Added: Direct internet access rules
Ingress: ec2.SecurityGroupIngressArray{
    // API port (8080)
    &ec2.SecurityGroupIngressArgs{
        Protocol:    pulumi.String("tcp"),
        FromPort:    pulumi.Int(8080),
        ToPort:      pulumi.Int(8080),
        CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
        Description: pulumi.String("API HTTP access"),
    },
    // SuperTokens port (3567)
    &ec2.SecurityGroupIngressArgs{
        Protocol:    pulumi.String("tcp"),
        FromPort:    pulumi.Int(3567),
        ToPort:      pulumi.Int(3567),
        CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
        Description: pulumi.String("SuperTokens HTTP access"),
    },
    // HTTP (80) and HTTPS (443) for optional frontend
    // ...
}
```

**Resource Struct Updates:**
```go
type AllInOneResources struct {
    Instance          *ec2.Instance
    SecurityGroup     *ec2.SecurityGroup
    PublicIP          pulumi.StringOutput  // Added
    PrivateIP         pulumi.StringOutput
    PublicDNS         pulumi.StringOutput  // Added
    // ...
}
```

**Function Signature:**
```go
// Before: Required ALB security group
func createAllInOneEC2(..., albSG *ec2.SecurityGroup, ...) 

// After: No ALB dependency
func createAllInOneEC2(..., masterUsername string, ...)
```

### 2. Main Pulumi Logic (`infra/main.go`)

**Removed:**
- `createAllInOneALB()` function call
- ALB-related exports (albDnsName, albArn, targetGroupArns)

**Added:**
```go
// Export instance access information
ctx.Export("allInOnePublicIp", allInOneRes.PublicIP)
ctx.Export("allInOnePublicDns", allInOneRes.PublicDNS)

// Service URLs with ports
ctx.Export("apiUrl", pulumi.Sprintf("http://%s:8080", allInOneRes.PublicDNS))
ctx.Export("apiPort", pulumi.String("8080"))
ctx.Export("supertokensUrl", pulumi.Sprintf("http://%s:3567", allInOneRes.PublicDNS))
ctx.Export("supertokensPort", pulumi.String("3567"))

// Connection info
ctx.Export("connectionInfo", pulumi.Sprintf(`
All-in-One Deployment - Direct Port Access:
- API: http://%s:8080
- SuperTokens: http://%s:3567
- Frontend (if deployed): http://%s:80
`, allInOneRes.PublicDNS, allInOneRes.PublicDNS, allInOneRes.PublicDNS))
```

### 3. Documentation Updates

**Updated Files:**
- `infra/ALLINONE_QUICKSTART.md`
  - Changed endpoint examples to use PUBLIC_DNS:PORT format
  - Updated cost breakdown (removed ALB and NAT Gateway)
  - Added pulumi stack output commands
  
- `infra/LOWCOST_ALLINONE.md`
  - Updated architecture diagrams
  - Added networking section explaining Internet Gateway vs NAT Gateway
  - Updated cost comparison table
  - Added benefits of direct port access

- `infra/ADMIN_INITIALIZATION.md`
  - Updated all curl examples to use PUBLIC_DNS:8080
  - Fixed API endpoint URLs

**New Files:**
- `infra/NETWORKING_ARCHITECTURE.md` - Comprehensive networking guide
- `infra/ALLINONE_NO_ALB_CHANGES.md` - This document

## Architecture Comparison

### Before (with ALB)

```
Internet
    ↓
Application Load Balancer ($16/mo)
    ├─ /api/* → Private Instance:8080
    └─ /auth/* → Private Instance:3567
    ↓
Private Subnet
    ↓
EC2 Instance (private IP only)
    ↓ (outbound)
NAT Gateway ($32/mo)
    ↓
Internet Gateway (FREE)
```

**Cost:** ~$58/month (ALB + NAT + EC2 + EBS)

### After (direct port access)

```
Internet
    ↓
Internet Gateway (FREE)
    ↓
Public Subnet
    ↓
EC2 Instance (public IP)
    ├─ Port 8080 → API
    ├─ Port 3567 → SuperTokens
    └─ Port 80 → Frontend (optional)
```

**Cost:** ~$10/month (EC2 + EBS only)

**Savings:** ~$48/month (83% reduction!)

## Networking Details

### Internet Gateway vs NAT Gateway

| Feature | Internet Gateway | NAT Gateway |
|---------|-----------------|-------------|
| Cost | FREE | $32/month |
| Use Case | Public subnets | Private subnets |
| Direction | Bidirectional | Outbound only |
| Instances | With public IPs | Without public IPs |

### Why No NAT Gateway?

**NAT Gateway is ONLY needed for:**
- Instances in private subnets
- That need outbound internet access
- But don't have public IPs

**Our all-in-one instance:**
- ✅ In public subnet
- ✅ Has public IP
- ✅ Uses Internet Gateway directly (FREE)
- ❌ Doesn't need NAT Gateway

## Service Access

### Before (via ALB)

```bash
# API
curl http://alb-dns-1234567.us-east-1.elb.amazonaws.com/api/health

# SuperTokens
curl http://alb-dns-1234567.us-east-1.elb.amazonaws.com/auth/hello
```

### After (direct ports)

```bash
# Get public DNS
PUBLIC_DNS=$(cd infra && pulumi stack output allInOnePublicDns)

# API (port 8080)
curl http://$PUBLIC_DNS:8080/health

# SuperTokens (port 3567)
curl http://$PUBLIC_DNS:3567/hello
```

## Security Considerations

### Concerns

**Public IP Exposure:**
- Instance is directly accessible from internet
- All open ports are exposed
- IP address is visible

### Mitigations

**Security Group Rules:**
```hcl
Ingress:
  - Port 8080: Allow from 0.0.0.0/0 (or restrict to specific IPs)
  - Port 3567: Allow from 0.0.0.0/0 (or restrict to frontend IPs)
  - Port 80/443: Allow from 0.0.0.0/0
  - Port 22: Can be removed (use SSM instead)

Egress:
  - All traffic: Allow (for updates, Docker pulls)
```

**Additional Security:**
- Use AWS Systems Manager Session Manager (no SSH needed)
- Enable CloudWatch logs
- Regular security updates
- Can add Fail2ban for brute force protection
- Optional: Restrict ports to specific IP ranges

### Production Recommendations

For production, consider:
1. **Use ALB mode** for better security and high availability
2. **Add CloudFront** in front for DDoS protection
3. **Restrict security group** to known IP ranges
4. **Enable AWS Shield** for DDoS protection
5. **Use WAF** for application-level protection
6. **SSL/TLS termination** (Let's Encrypt on instance)

## Migration Guide

### Upgrading Existing Deployment

**If you have an existing all-in-one deployment with ALB:**

```bash
# 1. Backup your data
# 2. Update Pulumi code (already done)
# 3. Run update
cd infra
pulumi up

# Pulumi will:
# - Move instance to public subnet
# - Assign public IP
# - Remove ALB and target groups
# - Update security group rules
```

**Update your frontend/clients:**

```javascript
// Before
const API_URL = 'http://alb-dns.elb.amazonaws.com/api';
const SUPERTOKENS_URL = 'http://alb-dns.elb.amazonaws.com';

// After
const PUBLIC_DNS = 'ec2-xx-xx-xx-xx.compute-1.amazonaws.com';
const API_URL = `http://${PUBLIC_DNS}:8080`;
const SUPERTOKENS_URL = `http://${PUBLIC_DNS}:3567`;
```

## Testing

### Verify Direct Port Access

```bash
# Get outputs
cd infra
PUBLIC_DNS=$(pulumi stack output allInOnePublicDns)
PUBLIC_IP=$(pulumi stack output allInOnePublicIp)

# Test API
curl -v http://$PUBLIC_DNS:8080/health
curl -v http://$PUBLIC_IP:8080/health

# Test SuperTokens
curl -v http://$PUBLIC_DNS:3567/hello

# Verify no ALB
pulumi stack output albDnsName
# Should return error or empty
```

### Security Group Verification

```bash
# List security group rules
INSTANCE_ID=$(pulumi stack output allInOneInstanceId)

aws ec2 describe-instances \
  --instance-ids $INSTANCE_ID \
  --query 'Reservations[0].Instances[0].SecurityGroups[0].GroupId' \
  --output text | xargs -I {} \
  aws ec2 describe-security-groups --group-ids {}
```

## Cost Breakdown

### Monthly Costs

**Infrastructure:**
- EC2 t3a.medium spot: $6-10/month
- EBS 30GB gp3: $2.40/month
- Data transfer: ~$1/month
- **Subtotal: ~$10/month**

**What You DON'T Pay:**
- ❌ ALB: Save $16/month
- ❌ NAT Gateway: Save $32/month
- ❌ ALB data processing: Save $5+/month
- ❌ NAT data processing: Save $5+/month

**Total Savings: ~$58/month (85% reduction!)**

### Annual Savings

- All-in-one: $120/year
- With ALB + NAT: $696/year
- **Savings: $576/year**

## Limitations

### What You Lose

1. **No path-based routing**
   - Can't route /api/* and /auth/* from same domain
   - Each service needs its own port

2. **No SSL termination at load balancer**
   - Must handle SSL on instance (Let's Encrypt)
   - Or use CloudFront in front

3. **No automatic high availability**
   - Single instance (can restart automatically)
   - No multi-AZ failover

4. **No automatic scaling**
   - Fixed instance size
   - Can manually scale up/down

5. **Security group instead of ALB security**
   - Less sophisticated rules
   - No WAF integration

### What You Keep

✅ All services still work
✅ Docker Compose orchestration
✅ Automatic container restarts
✅ Systems Manager access
✅ CloudWatch logging
✅ ECR integration
✅ Admin auto-initialization

## Rollback Procedure

If you need to add ALB back:

```bash
# 1. Keep the code changes
# 2. Add back ALB creation in main.go
# 3. Move instance back to private subnet
# 4. Update security group to allow ALB
# 5. Run pulumi up

# Or revert to previous commit
git revert HEAD
cd infra
pulumi up
```

## Future Enhancements

### Potential Improvements

1. **Add CloudFront:**
   - Put CloudFront in front of instance
   - SSL termination at edge
   - DDoS protection
   - Global caching
   - ~$1-5/month

2. **Let's Encrypt SSL:**
   - Add certbot to Docker Compose
   - Automatic SSL certificate
   - Update nginx to handle SSL

3. **Custom Domain:**
   - Point Route53 to public IP
   - Or use Elastic IP for static address
   - ~$3.60/month for Elastic IP

4. **Health Check Monitoring:**
   - CloudWatch alarm on port health
   - SNS notification on failures
   - Auto-restart container

## Related Documentation

- [ALLINONE_QUICKSTART.md](./ALLINONE_QUICKSTART.md) - Quick start guide
- [LOWCOST_ALLINONE.md](./LOWCOST_ALLINONE.md) - Architecture details
- [NETWORKING_ARCHITECTURE.md](./NETWORKING_ARCHITECTURE.md) - Networking deep dive
- [ADMIN_INITIALIZATION.md](./ADMIN_INITIALIZATION.md) - Admin setup

## Support

For issues or questions:
1. Check security group rules allow ports 8080, 3567
2. Verify instance has public IP assigned
3. Test from different network (avoid local firewall issues)
4. Check CloudWatch logs for instance startup
5. Connect via SSM to debug: `aws ssm start-session --target INSTANCE_ID`

---

**Implementation Date:** November 23, 2025  
**Status:** ✅ Complete and Tested  
**Breaking Changes:** Endpoint URLs changed from ALB DNS to PUBLIC_DNS:PORT  
**Cost Impact:** -$48-58/month (83-85% reduction)  
**Deployment Mode:** All-in-One only (Standard mode unchanged)

