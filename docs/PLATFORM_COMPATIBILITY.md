# Platform Compatibility Guide

**Last Updated**: November 24, 2025  
**Version**: 1.0

## Overview

The UTM Backend is designed to run on multiple architectures and platforms, with full support for both AMD64 (Intel/AMD) and ARM64 (Apple Silicon, AWS Graviton) processors.

## Supported Architectures

### ✅ AMD64 (x86_64)
- **Intel processors**: Core i3, i5, i7, i9, Xeon
- **AMD processors**: Ryzen, EPYC, Threadripper
- **Cloud platforms**: Most AWS EC2, Azure VMs, Google Cloud VMs
- **Desktop/Laptop**: Traditional Intel/AMD computers

### ✅ ARM64 (aarch64)
- **Apple Silicon**: M1, M1 Pro, M1 Max, M1 Ultra, M2, M2 Pro, M2 Max, M3, M3 Pro, M3 Max
- **Cloud platforms**: AWS Graviton (EC2, ECS, EKS), Azure Ampere, Google Cloud Tau
- **Server processors**: Ampere Altra, AWS Graviton2/3
- **Development boards**: Raspberry Pi 4/5 (8GB+ recommended)

## Docker Images

All services use multi-architecture Docker images:

| Service | Image | Architectures | Notes |
|---------|-------|---------------|-------|
| Nginx | `nginx:alpine` | AMD64, ARM64 | Official multi-arch |
| PostgreSQL | `postgres:16-alpine` | AMD64, ARM64 | Official multi-arch |
| Redis | `redis:7-alpine` | AMD64, ARM64 | Official multi-arch |
| SuperTokens | `supertokens/supertokens-postgresql:7.0` | AMD64, ARM64 | Official multi-arch |
| MailHog | `nfqlt/mailhog:latest` | AMD64, ARM64 | Community multi-arch ⭐ |
| API/Worker | Custom build | AMD64, ARM64 | Built from source |
| Frontend | Custom build | AMD64, ARM64 | Built from source |

### Why nfqlt/mailhog?

The official `mailhog/mailhog` image only supports AMD64. We use `nfqlt/mailhog` because:
- ✅ Supports both AMD64 and ARM64
- ✅ Actively maintained
- ✅ Drop-in replacement (same API/UI)
- ✅ Works identically to official image

## Platform-Specific Instructions

### Apple Silicon Macs (M1/M2/M3)

**No special configuration needed!** All images are ARM64 compatible.

```bash
# Standard setup works out of the box
docker-compose up -d
```

**Performance**: ARM64 native performance (faster than Rosetta 2 emulation)

**Verification**:
```bash
# Check architecture
docker ps --format "table {{.Names}}\t{{.Image}}" | while read name image; do
  if [ "$name" != "NAMES" ]; then
    arch=$(docker inspect "$name" --format='{{.Architecture}}' 2>/dev/null)
    echo "$name: $arch"
  fi
done
```

Should show `arm64` for all containers on Apple Silicon.

### Intel/AMD Macs

**No special configuration needed!** All images support AMD64.

```bash
docker-compose up -d
```

### Windows

#### Windows 11 (with WSL2)

**Recommended**: Use WSL2 with Docker Desktop

```bash
# In WSL2 Ubuntu/Debian
git clone <repository>
cd utm-backend
docker-compose up -d
```

**Architecture**: Depends on processor (AMD64 for Intel/AMD, ARM64 for ARM Windows devices)

#### Windows 10 (with WSL2)

Same as Windows 11. Ensure:
- WSL2 is enabled
- Docker Desktop is configured to use WSL2 backend
- Integration is enabled for your WSL2 distro

### Linux

#### x86_64 (AMD64)
```bash
docker-compose up -d
```

#### ARM64 (aarch64)
```bash
# Works on Raspberry Pi 4/5, ARM servers, etc.
docker-compose up -d
```

**Note for Raspberry Pi**: Minimum 8GB RAM recommended for full stack.

## Cloud Platforms

### AWS

#### EC2 Instances

**AMD64 instances**:
- t3, t3a, m5, m5a, c5, c5a, r5, r5a families
- Standard Ubuntu/Amazon Linux AMIs
- No special configuration needed

**ARM64 instances (Graviton)**:
- t4g, m6g, c6g, r6g, c7g families
- Use ARM64-compatible AMIs
- Cost savings: Up to 40% compared to x86 instances
- Better performance per dollar

```bash
# On Graviton instance
git clone <repository>
cd utm-backend
docker-compose up -d
```

#### ECS/Fargate

**AMD64**:
```json
{
  "cpu": "256",
  "memory": "512",
  "runtimePlatform": {
    "cpuArchitecture": "X86_64"
  }
}
```

**ARM64 (Graviton)**:
```json
{
  "cpu": "256",
  "memory": "512",
  "runtimePlatform": {
    "cpuArchitecture": "ARM64"
  }
}
```

### Google Cloud Platform

**AMD64**: Standard VMs (N1, N2, E2 series)  
**ARM64**: Tau T2A VMs (Ampere Altra processors)

Both work with standard docker-compose setup.

### Microsoft Azure

**AMD64**: Standard VMs (D, F, E series)  
**ARM64**: Limited availability (Ampere Altra based VMs)

### DigitalOcean

**AMD64**: All droplets  
**ARM64**: Not currently available

Use AMD64 droplets - all images compatible.

### Hetzner Cloud

**AMD64**: All cloud servers  
**ARM64**: ARM64 servers with CAX instances

Both architectures fully supported.

## Building from Source

### Multi-Architecture Build

Build for both architectures:

```bash
# Backend API/Worker
docker buildx build --platform linux/amd64,linux/arm64 -t utm-backend:latest .

# Frontend
cd frontend
docker buildx build --platform linux/amd64,linux/arm64 -t utm-frontend:latest .
```

### Single Architecture Build

Build for current platform only (faster):

```bash
# Builds for your current architecture
docker-compose build
```

## Performance Considerations

### ARM64 (Apple Silicon, Graviton)

**Advantages**:
- ✅ Better performance per watt (energy efficient)
- ✅ Lower costs on cloud platforms (AWS Graviton ~40% cheaper)
- ✅ Native performance (no emulation)
- ✅ Good memory bandwidth

**Considerations**:
- Some third-party tools may not have ARM64 builds
- Smaller ecosystem compared to AMD64

### AMD64 (Intel/AMD)

**Advantages**:
- ✅ Largest ecosystem
- ✅ Wide tool support
- ✅ More cloud instance options
- ✅ Better for legacy workloads

**Considerations**:
- Higher cloud costs compared to ARM64
- Higher power consumption

## Troubleshooting

### Issue: "exec format error"

**Cause**: Trying to run wrong architecture

**Solution**: Check Docker is using correct platform

```bash
# Check current platform
docker version | grep "OS/Arch"

# Force rebuild for correct platform
docker-compose build --no-cache
```

### Issue: MailHog won't start on Apple Silicon

**Symptom**:
```
utm-mailhog exited with code 1
WARNING: The requested image's platform (linux/amd64) does not match the detected host platform
```

**Solution**: Already fixed! We use `nfqlt/mailhog` which supports ARM64.

If you have the old image cached:
```bash
docker-compose down
docker rmi mailhog/mailhog
docker-compose pull mailhog
docker-compose up -d
```

### Issue: Slow performance on Apple Silicon

**Cause**: Running AMD64 images with Rosetta 2 emulation

**Check**:
```bash
docker ps --format "{{.Names}}: {{.Image}}"
docker inspect <container_name> --format='{{.Architecture}}'
```

**Solution**: Ensure all images are ARM64 native (see verification commands above)

### Issue: Image pull fails

**Symptom**:
```
no matching manifest for linux/arm64/v8
```

**Cause**: Image doesn't support your architecture

**Solution**: 
- For MailHog: Use `nfqlt/mailhog` (already configured)
- For other images: Check if multi-arch version exists
- Last resort: Run with emulation or use different service

## Architecture Detection

### Check System Architecture

```bash
# Linux/macOS
uname -m
# Output: x86_64 (AMD64) or aarch64/arm64 (ARM64)

# Check Docker architecture
docker version --format '{{.Server.Arch}}'

# Check image architecture
docker inspect postgres:16-alpine --format='{{.Architecture}}'
```

### Verify Multi-Arch Support

```bash
# Check available platforms for an image
docker buildx imagetools inspect postgres:16-alpine

# Expected output includes:
# - linux/amd64
# - linux/arm64
```

## CI/CD Considerations

### GitHub Actions

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        platform:
          - linux/amd64
          - linux/arm64
    steps:
      - uses: docker/setup-qemu-action@v2
      - uses: docker/setup-buildx-action@v2
      - name: Build
        run: |
          docker buildx build \
            --platform ${{ matrix.platform }} \
            -t utm-backend:${{ matrix.platform }} \
            --push .
```

### GitLab CI

```yaml
build:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker buildx create --use
    - docker buildx build --platform linux/amd64,linux/arm64 -t utm-backend:latest .
```

## Migration Guide

### From AMD64 to ARM64 (Cloud Cost Savings)

1. **Test locally** on ARM64 (Apple Silicon or ARM64 VM)
2. **Verify performance** meets requirements
3. **Update infrastructure** to ARM64 instances
4. **Deploy** using same docker-compose.yml
5. **Monitor** performance and costs

**Expected savings**: 30-40% on AWS Graviton vs comparable x86 instances

### From ARM64 to AMD64

No changes needed - all images support both architectures.

## Best Practices

### Development
✅ Use the architecture you'll deploy to  
✅ Test on multiple architectures before release  
✅ Use multi-arch images for all services  
✅ Keep Docker and Docker Compose updated

### Production
✅ Choose architecture based on cost/performance  
✅ ARM64 for cost savings (if workload compatible)  
✅ AMD64 for maximum compatibility  
✅ Use managed services where possible  
✅ Monitor resource usage and costs

### CI/CD
✅ Build for both architectures  
✅ Test on both architectures  
✅ Use buildx for multi-arch builds  
✅ Cache layers for faster builds

## Summary

**Current Status** ✅:
- All Docker images support both AMD64 and ARM64
- MailHog updated to multi-arch version (`nfqlt/mailhog`)
- Works on Apple Silicon without Rosetta 2 emulation
- Works on AWS Graviton for cost-optimized deployments
- Single docker-compose.yml works on all platforms

**No Platform-Specific Configuration Needed** 🎉

---

**Questions?**
- Check your architecture: `uname -m`
- Verify containers: `docker inspect <name> --format='{{.Architecture}}'`
- Report issues: Include output of `docker version` and `uname -m`


