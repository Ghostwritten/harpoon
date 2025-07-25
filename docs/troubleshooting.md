# Troubleshooting Guide

Common issues and solutions for Harpoon (hpn).

## Installation Issues

### Binary Not Found
```bash
Error: hpn: command not found
```

**Solutions:**
```bash
# Check if binary is in PATH
which hpn

# Add to PATH if needed
export PATH=$PATH:/usr/local/bin

# Or install to system location
sudo cp hpn /usr/local/bin/
sudo chmod +x /usr/local/bin/hpn
```

### Permission Denied
```bash
Error: permission denied
```

**Solutions:**
```bash
# Make binary executable
chmod +x hpn

# Or run with sudo
sudo ./hpn -a pull -f images.txt

# Add user to docker group (for Docker runtime)
sudo usermod -aG docker $USER
newgrp docker
```

## Runtime Issues

### No Container Runtime Found
```bash
Error: no container runtime found. Please install docker, podman, or nerdctl
```

**Solutions:**
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Or install Podman
sudo apt-get install podman  # Ubuntu/Debian
brew install podman          # macOS

# Or install Nerdctl
# See: https://github.com/containerd/nerdctl
```

### Runtime Not Available
```bash
Error: specified runtime 'docker' is not available
```

**Solutions:**
```bash
# Check if Docker daemon is running
sudo systemctl status docker
sudo systemctl start docker

# Or use auto-fallback
hpn --auto-fallback -a pull -f images.txt

# Or specify different runtime
hpn --runtime podman -a pull -f images.txt
```

### Docker Permission Issues
```bash
Error: permission denied while trying to connect to Docker daemon
```

**Solutions:**
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Or use sudo
sudo hpn -a pull -f images.txt

# Or use Podman (rootless)
hpn --runtime podman -a pull -f images.txt
```

## Network Issues

### Registry Connection Failed
```bash
Error: failed to pull image nginx:latest: connection refused
```

**Solutions:**
```bash
# Check network connectivity
ping registry.company.com

# Check DNS resolution
nslookup registry.company.com

# Test with curl
curl -I https://registry.company.com/v2/

# Check proxy settings
echo $http_proxy
echo $https_proxy
```

### Proxy Configuration Issues
```bash
Error: proxyconnect tcp: dial tcp proxy:8080: connection refused
```

**Solutions:**
```bash
# Set proxy environment variables
export http_proxy=http://proxy.company.com:8080
export https_proxy=http://proxy.company.com:8080

# Or configure in config file
cat > ~/.hpn/config.yaml << EOF
proxy:
  http: http://proxy.company.com:8080
  https: http://proxy.company.com:8080
  enabled: true
EOF

# Test proxy connectivity
curl -x http://proxy.company.com:8080 http://google.com
```

### SSL/TLS Certificate Issues
```bash
Error: x509: certificate signed by unknown authority
```

**Solutions:**
```bash
# Add CA certificate to system trust store
sudo cp company-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Or configure Docker to trust registry
sudo mkdir -p /etc/docker/certs.d/registry.company.com
sudo cp company-ca.crt /etc/docker/certs.d/registry.company.com/ca.crt

# Or use insecure registry (not recommended for production)
# Add to Docker daemon.json:
{
  "insecure-registries": ["registry.company.com"]
}
```

## Authentication Issues

### Registry Authentication Failed
```bash
Error: authentication failed for registry.company.com
```

**Solutions:**
```bash
# Login to registry
docker login registry.company.com

# Check credentials
docker logout registry.company.com
docker login registry.company.com

# Use token authentication
echo "your-token" | docker login registry.company.com -u username --password-stdin

# Check credential helper
docker-credential-desktop list
```

### Token Expired
```bash
Error: unauthorized: authentication required
```

**Solutions:**
```bash
# Re-login to registry
docker logout registry.company.com
docker login registry.company.com

# Check token expiration
# Tokens typically expire after 12-24 hours

# Use long-lived service account
# Create service account in registry UI
```

## File System Issues

### Insufficient Disk Space
```bash
Error: no space left on device
```

**Solutions:**
```bash
# Check disk space
df -h

# Clean up Docker images
docker system prune -a

# Clean up old tar files
find ./images -name "*.tar" -mtime +7 -delete

# Use different directory
hpn -a save -f images.txt --save-mode 2
# Then move images/ to different disk
```

### File Permission Issues
```bash
Error: permission denied: ./images/nginx_latest.tar
```

**Solutions:**
```bash
# Fix directory permissions
chmod 755 ./images/
chmod 644 ./images/*.tar

# Change ownership
sudo chown -R $USER:$USER ./images/

# Create directory with correct permissions
mkdir -p ./images
chmod 755 ./images
```

### File Not Found
```bash
Error: failed to read image list: no such file or directory
```

**Solutions:**
```bash
# Check file exists
ls -la images.txt

# Check file path
pwd
realpath images.txt

# Create sample image list
cat > images.txt << EOF
nginx:latest
alpine:3.18
EOF
```

## Configuration Issues

### Invalid Configuration
```bash
Error: invalid log level 'verbose'. Valid values: debug, info, warn, error
```

**Solutions:**
```bash
# Check configuration syntax
hpn --config ~/.hpn/config.yaml -a pull -f /dev/null

# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('~/.hpn/config.yaml'))"

# Use default configuration
mv ~/.hpn/config.yaml ~/.hpn/config.yaml.backup
```

### Configuration Not Found
```bash
Error: config file not found: /path/to/config.yaml
```

**Solutions:**
```bash
# Check file exists
ls -la /path/to/config.yaml

# Use default locations
mkdir -p ~/.hpn
cp config.yaml.example ~/.hpn/config.yaml

# Specify correct path
hpn --config ./config.yaml -a pull -f images.txt
```

## Parameter Validation Issues

### Invalid Mode Parameter
```bash
Error: --save-mode cannot be used with push action
```

**Solutions:**
```bash
# Use correct mode parameter for action
hpn -a push -f images.txt --push-mode 2    # Correct
hpn -a save -f images.txt --save-mode 2    # Correct

# Check help for valid parameters
hpn --help
```

### Invalid Mode Value
```bash
Error: invalid push-mode '3'. Valid values: 1, 2
```

**Solutions:**
```bash
# Use valid mode values
hpn -a push -f images.txt --push-mode 1    # Simple push
hpn -a push -f images.txt --push-mode 2    # With project

# Check documentation for mode descriptions
hpn --help
```

## Performance Issues

### Slow Image Operations
```bash
# Operations taking too long
```

**Solutions:**
```bash
# Increase parallel workers
cat > ~/.hpn/config.yaml << EOF
parallel:
  max_workers: 8
  auto_adjust: true
EOF

# Use local registry mirror
registry: registry-mirror.company.com

# Optimize network settings
runtime:
  timeout: 15m
  retry:
    max_attempts: 3
    delay: 1s
```

### Memory Issues
```bash
Error: out of memory
```

**Solutions:**
```bash
# Reduce parallel workers
parallel:
  max_workers: 2
  auto_adjust: false

# Process images in batches
split -l 10 large-image-list.txt batch-
for batch in batch-*; do
  hpn -a pull -f "$batch"
done

# Increase system memory or swap
```

## Debug Mode

### Enable Debug Logging
```bash
# Environment variable
export HPN_LOG_LEVEL=debug
hpn -a pull -f images.txt

# Configuration file
logging:
  level: debug
  console: true
  file: ./debug.log

# Analyze debug output
grep -i error debug.log
grep -i timeout debug.log
grep -i retry debug.log
```

### Verbose Output
```bash
# Enable verbose mode
hpn -a pull -f images.txt 2>&1 | tee verbose.log

# Check specific operations
grep "Pulling" verbose.log
grep "Using runtime" verbose.log
grep "Summary" verbose.log
```

## Common Error Patterns

### Network Timeouts
```bash
Error: context deadline exceeded
Error: i/o timeout
```

**Solutions:**
```bash
# Increase timeout
runtime:
  timeout: 15m

# Check network connectivity
ping registry.company.com
traceroute registry.company.com

# Use retry mechanism
runtime:
  retry:
    max_attempts: 5
    delay: 2s
    max_delay: 60s
```

### Image Not Found
```bash
Error: image not found: nginx:nonexistent
Error: manifest unknown
```

**Solutions:**
```bash
# Check image name and tag
docker search nginx
curl -s https://registry.hub.docker.com/v2/repositories/library/nginx/tags/

# Use correct registry
hpn -a pull -f images.txt -r docker.io

# Check image list file
cat images.txt
```

### Resource Exhaustion
```bash
Error: too many open files
Error: resource temporarily unavailable
```

**Solutions:**
```bash
# Increase file descriptor limit
ulimit -n 4096

# Reduce parallel workers
parallel:
  max_workers: 4

# Check system resources
free -h
df -h
```

## Getting Help

### Command Help
```bash
# General help
hpn --help

# Version information
hpn --version
hpn version

# Check configuration
hpn --config ~/.hpn/config.yaml --help
```

### Log Analysis
```bash
# Check system logs
journalctl -u docker
tail -f /var/log/docker.log

# Check hpn logs
tail -f ~/.hpn/hpn.log
grep -i error ~/.hpn/hpn.log
```

### System Information
```bash
# Check system info
uname -a
docker version
podman version

# Check Go version (if building from source)
go version

# Check network configuration
ip route
cat /etc/resolv.conf
```

## Reporting Issues

### Information to Include
When reporting issues, include:

1. **Version Information**:
   ```bash
   hpn version
   ```

2. **System Information**:
   ```bash
   uname -a
   docker version
   ```

3. **Configuration**:
   ```bash
   cat ~/.hpn/config.yaml
   ```

4. **Error Output**:
   ```bash
   export HPN_LOG_LEVEL=debug
   hpn -a pull -f images.txt 2>&1 | tee error.log
   ```

5. **Steps to Reproduce**:
   - Exact commands used
   - Image list contents
   - Expected vs actual behavior

### GitHub Issues
- [Report Bug](https://github.com/your-org/harpoon/issues/new?template=bug_report.md)
- [Feature Request](https://github.com/your-org/harpoon/issues/new?template=feature_request.md)
- [Question](https://github.com/your-org/harpoon/discussions)

## See Also

- [User Guide](user-guide.md) - Complete usage guide
- [Configuration](configuration.md) - Configuration reference
- [Examples](examples.md) - Real-world examples
- [Security](security.md) - Security best practices