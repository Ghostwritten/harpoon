# Security Guide

Security best practices and configuration for Harpoon (hpn).

## GitHub Actions Security

### Required Secrets

For secure GitHub Actions testing, configure these secrets in your repository:

**Path**: Repository Settings → Secrets and variables → Actions

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `DOCKER_HUB_USERNAME` | Docker Hub username | `your-username` |
| `DOCKER_HUB_TOKEN` | Docker Hub access token | `dckr_pat_...` |

### Secret Configuration Steps

1. **Generate Docker Hub Token**:
   - Login to Docker Hub
   - Go to Account Settings → Security
   - Create New Access Token
   - Set permissions: Read, Write, Delete

2. **Configure GitHub Secrets**:
   - Go to GitHub repository
   - Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Add the required secrets

### Security Best Practices

#### Token Permissions
```yaml
# Recommended Docker Hub token permissions
permissions:
  read: true      # Pull images
  write: true     # Push test images
  delete: true    # Clean up test images
  admin: false    # Not needed
```

#### Environment Isolation
```yaml
# GitHub Actions environment protection
environment:
  name: testing
  protection_rules:
    - required_reviewers: 1
    - wait_timer: 5
```

#### Branch Protection
```yaml
# Limit sensitive operations to specific branches
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
```

## Code Security

### Sensitive Information Detection

Check for hardcoded secrets:
```bash
# Search for potential secrets
grep -r "token\|password\|secret\|key" . --exclude-dir=.git
grep -r "dckr_pat_\|ghp_" . --exclude-dir=.git

# Check configuration files
find . -name "*.yaml" -o -name "*.yml" | xargs grep -l "token\|password"
```

### Git History Cleanup

If secrets were accidentally committed:
```bash
# Remove sensitive files from Git history
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch path/to/sensitive/file' \
  --prune-empty --tag-name-filter cat -- --all

# Alternative: use BFG Repo-Cleaner
java -jar bfg.jar --delete-files sensitive-file.txt
git reflog expire --expire=now --all && git gc --prune=now --aggressive
```

### Pre-commit Hooks

Create `.pre-commit-config.yaml`:
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: detect-private-key
      - id: check-added-large-files
      - id: check-merge-conflict
  
  - repo: https://github.com/Yelp/detect-secrets
    rev: v1.4.0
    hooks:
      - id: detect-secrets
        args: ['--baseline', '.secrets.baseline']
```

## Runtime Security

### Container Runtime Security

#### Docker Security
```bash
# Run as non-root user
docker run --user $(id -u):$(id -g) ...

# Use read-only root filesystem
docker run --read-only ...

# Limit resources
docker run --memory=512m --cpus=1.0 ...
```

#### Podman Security
```bash
# Rootless containers (default in Podman)
podman run --user $(id -u):$(id -g) ...

# Use security profiles
podman run --security-opt seccomp=default.json ...
```

### Image Security

#### Image Verification
```bash
# Verify image signatures (if available)
docker trust inspect image:tag

# Check image vulnerabilities
docker scan image:tag

# Use specific tags, avoid 'latest'
hpn -a pull -f images.txt  # where images.txt contains specific tags
```

#### Secure Image Sources
```bash
# Use trusted registries
registry: harbor.company.com

# Verify image sources
# Only pull from known, trusted registries
```

## Network Security

### Proxy Configuration

#### Corporate Proxy
```yaml
# config.yaml
proxy:
  http: http://proxy.company.com:8080
  https: http://proxy.company.com:8080
  enabled: true
```

#### Environment Variables
```bash
export http_proxy=http://proxy:8080
export https_proxy=http://proxy:8080
export no_proxy=localhost,127.0.0.1,.company.com
```

### Registry Authentication

#### Secure Authentication
```bash
# Use Docker credential helpers
docker login registry.company.com

# Store credentials securely
# ~/.docker/config.json should use credential helpers
{
  "credHelpers": {
    "registry.company.com": "desktop"
  }
}
```

#### Token-based Authentication
```bash
# Use tokens instead of passwords
docker login -u username --password-stdin registry.com < token.txt
```

## Configuration Security

### Secure Configuration

#### File Permissions
```bash
# Secure config file permissions
chmod 600 ~/.hpn/config.yaml
chown $(whoami):$(whoami) ~/.hpn/config.yaml
```

#### Environment Variables
```bash
# Use environment variables for sensitive data
export HPN_REGISTRY_TOKEN="$(cat /secure/path/token)"
export HPN_PROXY_PASSWORD="$(cat /secure/path/proxy-pass)"
```

### Configuration Validation

#### Input Validation
```yaml
# Validate registry URLs
registry: https://registry.company.com  # Use HTTPS

# Validate proxy URLs
proxy:
  http: http://proxy.company.com:8080   # Explicit protocol
  https: http://proxy.company.com:8080
```

#### Security Checks
```bash
# Check configuration security
hpn config validate --security-check

# Audit configuration
hpn config audit
```

## Incident Response

### Token Compromise

If a token is compromised:

1. **Immediate Actions**:
   ```bash
   # Revoke the compromised token
   # Login to Docker Hub → Account Settings → Security
   # Find and delete the compromised token
   ```

2. **Generate New Token**:
   ```bash
   # Create new access token with minimal permissions
   # Update GitHub Secrets with new token
   ```

3. **Clean Git History**:
   ```bash
   # If token was committed to Git
   git filter-branch --force --index-filter \
     'git rm --cached --ignore-unmatch file-with-token' \
     --prune-empty --tag-name-filter cat -- --all
   ```

4. **Audit Access**:
   ```bash
   # Check Docker Hub access logs
   # Review recent image pulls/pushes
   # Verify no unauthorized access
   ```

### Security Monitoring

#### Log Monitoring
```bash
# Monitor for suspicious activities
grep -i "auth\|login\|token" /var/log/hpn.log

# Check for unusual registry access
grep -i "registry\|push\|pull" /var/log/hpn.log
```

#### Alerting
```yaml
# GitHub Actions notification on security events
- name: Security Alert
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: failure
    webhook_url: ${{ secrets.SECURITY_WEBHOOK }}
    text: "Security test failed in ${{ github.repository }}"
```

## Compliance

### Security Standards

#### NIST Guidelines
- Use strong authentication (multi-factor when possible)
- Implement least privilege access
- Monitor and log security events
- Regular security assessments

#### Industry Best Practices
- Regular token rotation
- Secure credential storage
- Network segmentation
- Vulnerability scanning

### Audit Trail

#### Logging Requirements
```yaml
# Comprehensive logging
logging:
  level: info
  format: json
  file: /var/log/hpn-audit.log
  console: false
  timestamp: true
```

#### Audit Events
- Authentication attempts
- Registry access
- Configuration changes
- Error conditions
- Administrative actions

## Security Checklist

### Development
- [ ] No hardcoded secrets in code
- [ ] Pre-commit hooks configured
- [ ] Dependency vulnerability scanning
- [ ] Code review for security issues

### Deployment
- [ ] GitHub Secrets properly configured
- [ ] Token permissions minimized
- [ ] Branch protection rules enabled
- [ ] Environment protection configured

### Operations
- [ ] Regular token rotation
- [ ] Security monitoring enabled
- [ ] Incident response plan ready
- [ ] Audit logs reviewed regularly

### Configuration
- [ ] Secure file permissions
- [ ] HTTPS for all registry URLs
- [ ] Proxy configuration secured
- [ ] Input validation enabled

## Resources

### Security Tools
- [GitHub Security Advisories](https://github.com/advisories)
- [Docker Security Scanning](https://docs.docker.com/engine/scan/)
- [OWASP Container Security](https://owasp.org/www-project-container-security/)

### Documentation
- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [Container Security Guide](https://kubernetes.io/docs/concepts/security/)

### Compliance
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [CIS Controls](https://www.cisecurity.org/controls/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)