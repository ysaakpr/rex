# Low-Cost Mode - Quick Start

## 🚀 Enable Low-Cost Mode in 3 Steps

### 1. Configure
```bash
cd infra
pulumi config set rex-backend:lowcost true
```

### 2. Deploy
```bash
pulumi up
```

### 3. Verify
```bash
pulumi stack output deploymentMode
# Should show: "lowcost"

pulumi stack output lowCostInstancePrivateIp
# Shows the EC2 instance IP
```

## 💰 Cost Savings
- **Before**: ~$65-130/month (RDS + ElastiCache)
- **After**: ~$5-10/month (EC2 Spot)
- **Savings**: 85-90% 🎉

## 📋 What Gets Deployed

```
Single EC2 Spot Instance (t3a.small)
├─ PostgreSQL 14
│  ├─ rex_backend database
│  └─ supertokens database
└─ Redis Server
   └─ 512 MB max memory
```

## 🔧 Common Commands

### Access the Instance
```bash
# Get instance ID
INSTANCE_ID=$(pulumi stack output lowCostInstanceId)

# Connect (no SSH key needed!)
aws ssm start-session --target $INSTANCE_ID
```

### Check Services
```bash
# Once connected to instance:

# PostgreSQL status
sudo systemctl status postgresql

# Redis status
sudo systemctl status redis-server

# Test PostgreSQL
sudo -u postgres psql -l

# Test Redis
redis-cli ping
```

### Switch Back to Standard Mode
```bash
pulumi config set rex-backend:lowcost false
pulumi up
```

## ⚠️ Important Notes

### ✅ Perfect For:
- Development environments
- Testing/staging
- Learning projects
- Personal projects

### ❌ Not For:
- Production systems (use standard mode)
- High-availability requirements
- Mission-critical applications

## 🐛 Troubleshooting

### Can't connect to database?
```bash
# Check if instance is running
aws ec2 describe-instances --instance-ids $(pulumi stack output lowCostInstanceId)

# Connect and check logs
aws ssm start-session --target $(pulumi stack output lowCostInstanceId)
sudo journalctl -u postgresql -n 50
```

### Spot instance interrupted?
```bash
# Check status
aws ec2 describe-instance-status --instance-ids $(pulumi stack output lowCostInstanceId)

# Restart if needed (usually automatic)
aws ec2 start-instances --instance-ids $(pulumi stack output lowCostInstanceId)
```

## 📚 More Information

- **Complete Guide**: [LOWCOST_MODE.md](./LOWCOST_MODE.md)
- **Implementation Details**: [LOWCOST_IMPLEMENTATION_SUMMARY.md](./LOWCOST_IMPLEMENTATION_SUMMARY.md)
- **Main README**: [README.md](./README.md)

## 💡 Pro Tips

1. **Backup regularly**: Use EBS snapshots or pg_dump
2. **Monitor costs**: Check AWS Cost Explorer after 24 hours
3. **Test recovery**: Practice restoring from backups
4. **Use SSM**: No need for SSH keys or bastion hosts
5. **Check logs**: CloudWatch Logs show instance initialization

## 🎯 Next Steps

1. ✅ Deploy with low-cost mode
2. ✅ Run your application
3. ✅ Monitor costs (should drop dramatically)
4. ✅ Set up backups (EBS snapshots recommended)
5. ✅ Enjoy the savings! 💰

---

**Questions?** See [LOWCOST_MODE.md](./LOWCOST_MODE.md) for detailed documentation.

