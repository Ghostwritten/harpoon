# Upgrade Guide

Guide for upgrading between Harpoon (hpn) versions.

## Upgrading to v1.1

### Overview

Harpoon v1.1 includes breaking changes, primarily related to push mode simplification. This guide helps you upgrade smoothly.

### Breaking Changes

#### 1. Push Mode 3 Removed

**Change**: Removed Push Mode 3 (preserve original project path)

**Impact**: Scripts using `--push-mode 3` need updates

**Migration**:
```bash
# v1.0 (old version)
hpn -a push -f images.txt -r registry.com --push-mode 3

# v1.1 (new version) - equivalent operation
hpn -a push -f images.txt -r registry.com --push-mode 2
# Note: Don't specify -p parameter, let system use original image project name
```

#### 2. Push Mode 2 Behavior Change

**Change**: Push Mode 2 now uses smart project name selection

**New Priority**:
1. Command line parameter (`-p project`)
2. Configuration file (`project: name`)
3. Original image project name

**Migration Check**:
```bash
# Check if your scripts depend on old Mode 2 behavior
grep -r "push-mode 2" your-scripts/
grep -r "push.*-p" your-scripts/
```

### Configuration Updates

#### New Configuration Options

Add to your `~/.hpn/config.yaml`:

```yaml
runtime:
  preferred: docker          # Options: docker, podman, nerdctl
  auto_fallback: false      # Auto fallback to other runtimes
  timeout: 5m               # Runtime operation timeout
  retry:
    max_attempts: 3
    delay: 1s
    max_delay: 30s
```

#### Push Mode Configuration

```yaml
modes:
  push_mode: 1              # 1=simple push, 2=smart project push
  # Note: push_mode no longer supports value 3
```

### Script Migration Examples

#### Scenario 1: Scripts Using Push Mode 3

**Old Script**:
```bash
#!/bin/bash
# Preserve original project structure
hpn -a push -f production-images.txt -r harbor.company.com --push-mode 3
```

**New Script**:
```bash
#!/bin/bash
# Use smart project selection (equivalent to old Mode 3)
hpn -a push -f production-images.txt -r harbor.company.com --push-mode 2
```

#### Scenario 2: Scripts Specifying Specific Project

**Old Script**:
```bash
#!/bin/bash
# Push to specific project
hpn -a push -f images.txt -r registry.com -p myproject --push-mode 2
```

**New Script**:
```bash
#!/bin/bash
# Behavior unchanged, but logic is smarter
hpn -a push -f images.txt -r registry.com -p myproject --push-mode 2
```

#### Scenario 3: CI/CD Pipeline Updates

**Old CI Configuration**:
```yaml
script:
  - hpn -a push -f images.txt --push-mode 3 -r $REGISTRY
```

**New CI Configuration**:
```yaml
script:
  - hpn -a push -f images.txt --push-mode 2 -r $REGISTRY --auto-fallback
  # Added --auto-fallback for automatic runtime fallback in CI
```

### Verification

#### Version Check
```bash
hpn -v
# Should display v1.1
```

#### Functionality Test
```bash
# Test runtime detection
hpn --runtime docker -a pull -f test-images.txt

# Test new push mode
hpn -a push -f test-images.txt -r localhost:5000 --push-mode 2

# Test parameter validation
hpn -a push -f test-images.txt --save-mode 2
# Should display error: --save-mode cannot be used with push action
```

#### Configuration Validation
```bash
# Check configuration loading
hpn -a pull -f test-images.txt --config ~/.hpn/config.yaml
```

### Common Issues

#### Q: My script uses `--push-mode 3`, what happens after upgrade?
A: You'll get error: `invalid push-mode '3'. Valid values: 1, 2`. Change to `--push-mode 2`.

#### Q: Will Push Mode 2's new behavior affect my existing pushes?
A: If you explicitly specify `-p` parameter, behavior remains unchanged. Without `-p`, it now uses original image project name.

#### Q: How to handle runtime unavailability in CI?
A: Use `--auto-fallback` parameter, or set `runtime.auto_fallback: true` in config file.

#### Q: Error message format changed?
A: Yes, error messages are now more concise and don't show full help information.

### Rollback Plan

If you encounter issues after upgrade, you can temporarily rollback to v1.0:

```bash
# Rebuild v1.0
git checkout v1.0
go build -o hpn-v1.0 ./cmd/hpn

# Or keep both versions
mv hpn hpn-v1.1
mv hpn-v1.0 hpn
```

### Getting Help

If you encounter issues during upgrade:

1. Check [Changelog](changelog.md) for detailed changes
2. Review [User Guide](user-guide.md) for updated usage
3. Submit GitHub Issue for support

### Upgrade Checklist

- [ ] Backup existing scripts and configuration
- [ ] Update version to v1.1
- [ ] Search and update all scripts using `--push-mode 3`
- [ ] Update configuration file with new runtime options
- [ ] Test critical push operations
- [ ] Verify CI/CD pipelines work correctly
- [ ] Update documentation and team training materials

## Upgrading from Earlier Versions

### From v0.x to v1.x

This is a major version upgrade with significant changes:

#### Key Changes
- Complete rewrite in Go (from shell script)
- New command-line interface
- Configuration file format change
- Enhanced runtime support

#### Migration Steps
1. **Backup existing configuration**
2. **Install new binary**
3. **Convert configuration format**
4. **Update all scripts**
5. **Test thoroughly**

#### Configuration Migration
```bash
# Old shell script configuration
REGISTRY="harbor.company.com"
PROJECT="production"

# New YAML configuration
cat > ~/.hpn/config.yaml << EOF
registry: harbor.company.com
project: production
runtime:
  preferred: docker
modes:
  push_mode: 2
EOF
```

#### Script Migration
```bash
# Old shell script usage
./images.sh -a pull -f images.txt

# New Go binary usage
hpn -a pull -f images.txt
```

## Version Compatibility Matrix

| Feature | v0.x | v1.0 | v1.1 |
|---------|------|------|------|
| Shell Script | ✅ | ❌ | ❌ |
| Go Binary | ❌ | ✅ | ✅ |
| Push Mode 3 | ❌ | ✅ | ❌ |
| Smart Project Selection | ❌ | ❌ | ✅ |
| Runtime Auto-fallback | ❌ | ❌ | ✅ |
| Parameter Validation | ❌ | ❌ | ✅ |

## Best Practices for Upgrades

### Testing Strategy
1. **Test in development environment first**
2. **Use feature branches for script updates**
3. **Run comprehensive tests before production**
4. **Have rollback plan ready**

### Gradual Migration
1. **Update non-critical scripts first**
2. **Monitor for issues**
3. **Update critical production scripts last**
4. **Document all changes**

### Team Communication
1. **Notify team of breaking changes**
2. **Provide migration timeline**
3. **Share updated documentation**
4. **Conduct training sessions if needed**

## See Also

- [Changelog](changelog.md) - Complete version history
- [User Guide](user-guide.md) - Updated usage guide
- [Configuration](configuration.md) - New configuration options
- [Troubleshooting](troubleshooting.md) - Common upgrade issues