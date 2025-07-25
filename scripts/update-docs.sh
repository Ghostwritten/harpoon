#!/bin/bash

################################################################################
# Documentation Update Script
# 自动更新项目文档的脚本
################################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 默认值
UPDATE_TYPE="all"
DRY_RUN=false
FORCE=false
OUTPUT_DIR="docs"

# 显示帮助信息
show_help() {
    echo "Documentation Update Script"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -t, --type TYPE         更新类型 (api|cli|user|changelog|all)"
    echo "  -o, --output DIR        输出目录 (默认: docs)"
    echo "  -d, --dry-run           预览模式，不写入文件"
    echo "  -f, --force             强制更新，覆盖现有文件"
    echo "  -h, --help              显示帮助信息"
    echo ""
    echo "更新类型:"
    echo "  api         更新API文档"
    echo "  cli         更新CLI参考文档"
    echo "  user        更新用户指南"
    echo "  changelog   更新变更日志"
    echo "  all         更新所有文档"
    echo ""
    echo "示例:"
    echo "  $0 -t api                # 只更新API文档"
    echo "  $0 -t all --dry-run      # 预览所有文档更新"
    echo "  $0 -f                    # 强制更新所有文档"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            UPDATE_TYPE="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            print_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证参数
validate_parameters() {
    case "$UPDATE_TYPE" in
        "api"|"cli"|"user"|"changelog"|"all")
            ;;
        *)
            print_error "无效的更新类型: $UPDATE_TYPE"
            exit 1
            ;;
    esac
    
    if [ ! -d "$OUTPUT_DIR" ]; then
        mkdir -p "$OUTPUT_DIR"
        print_info "创建输出目录: $OUTPUT_DIR"
    fi
}

# 检查先决条件
check_prerequisites() {
    print_info "检查先决条件..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go未安装或不在PATH中"
        exit 1
    fi
    
    # 检查是否在项目根目录
    if [ ! -f "go.mod" ]; then
        print_error "请在项目根目录运行此脚本"
        exit 1
    fi
    
    # 构建应用
    print_info "构建应用..."
    go build -o hpn ./cmd/hpn
    
    print_success "先决条件检查通过"
}

# 更新API文档
update_api_docs() {
    print_info "更新API文档..."
    
    local api_doc="$OUTPUT_DIR/api-reference.md"
    
    if [ "$DRY_RUN" = true ]; then
        print_info "预览模式: 将生成 $api_doc"
        return 0
    fi
    
    # 检查是否覆盖现有文件
    if [ -f "$api_doc" ] && [ "$FORCE" = false ]; then
        read -p "文件 $api_doc 已存在，是否覆盖？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "跳过API文档更新"
            return 0
        fi
    fi
    
    # 生成API文档
    cat > "$api_doc" << 'EOF'
# API Reference

This document provides detailed API reference for Harpoon.

## Overview

Harpoon provides a comprehensive API for container image management operations.

EOF
    
    # 添加包文档
    echo "## Packages" >> "$api_doc"
    echo "" >> "$api_doc"
    
    # 遍历主要包
    for pkg_dir in cmd/hpn pkg/*/; do
        if [ -d "$pkg_dir" ] && [ -n "$(find "$pkg_dir" -name "*.go" -not -name "*_test.go")" ]; then
            pkg_name=$(basename "$pkg_dir")
            echo "### Package: $pkg_name" >> "$api_doc"
            echo "" >> "$api_doc"
            
            # 获取包文档
            if go doc "./$pkg_dir" &>/dev/null; then
                go doc "./$pkg_dir" | head -10 >> "$api_doc"
            else
                echo "No documentation available for package $pkg_name." >> "$api_doc"
            fi
            echo "" >> "$api_doc"
        fi
    done
    
    print_success "API文档已更新: $api_doc"
}

# 更新CLI文档
update_cli_docs() {
    print_info "更新CLI文档..."
    
    local cli_doc="$OUTPUT_DIR/cli-reference.md"
    
    if [ "$DRY_RUN" = true ]; then
        print_info "预览模式: 将生成 $cli_doc"
        return 0
    fi
    
    # 检查是否覆盖现有文件
    if [ -f "$cli_doc" ] && [ "$FORCE" = false ]; then
        read -p "文件 $cli_doc 已存在，是否覆盖？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "跳过CLI文档更新"
            return 0
        fi
    fi
    
    # 生成CLI文档
    cat > "$cli_doc" << 'EOF'
# Command Line Interface Reference

This document provides detailed CLI reference for Harpoon.

## Overview

Harpoon provides a comprehensive command-line interface for container image management.

EOF
    
    # 添加主命令帮助
    echo "## Main Commands" >> "$cli_doc"
    echo "" >> "$cli_doc"
    echo '```' >> "$cli_doc"
    ./hpn --help >> "$cli_doc" 2>/dev/null || echo "Help not available" >> "$cli_doc"
    echo '```' >> "$cli_doc"
    echo "" >> "$cli_doc"
    
    # 添加版本信息
    echo "## Version Information" >> "$cli_doc"
    echo "" >> "$cli_doc"
    echo '```' >> "$cli_doc"
    ./hpn --version >> "$cli_doc" 2>/dev/null || echo "Version not available" >> "$cli_doc"
    echo '```' >> "$cli_doc"
    echo "" >> "$cli_doc"
    
    # 添加子命令文档
    echo "## Subcommands" >> "$cli_doc"
    echo "" >> "$cli_doc"
    
    # 尝试获取子命令列表
    subcommands=("pull" "push" "save" "load" "version")
    for cmd in "${subcommands[@]}"; do
        echo "### $cmd" >> "$cli_doc"
        echo "" >> "$cli_doc"
        echo '```' >> "$cli_doc"
        ./hpn "$cmd" --help >> "$cli_doc" 2>/dev/null || echo "Help not available for $cmd" >> "$cli_doc"
        echo '```' >> "$cli_doc"
        echo "" >> "$cli_doc"
    done
    
    print_success "CLI文档已更新: $cli_doc"
}

# 更新用户指南
update_user_guide() {
    print_info "更新用户指南..."
    
    local user_doc="$OUTPUT_DIR/user-guide.md"
    
    if [ "$DRY_RUN" = true ]; then
        print_info "预览模式: 将更新 $user_doc"
        return 0
    fi
    
    # 如果文件不存在，创建基础模板
    if [ ! -f "$user_doc" ]; then
        cat > "$user_doc" << 'EOF'
# User Guide

This guide provides comprehensive information on using Harpoon for container image management.

## Getting Started

### Installation

Please refer to the [Installation Guide](installation.md) for detailed installation instructions.

### Quick Start

Here are some basic examples to get you started:

```bash
# Pull images
hpn -a pull -f images.txt

# Save images
hpn -a save -f images.txt --save-mode 2

# Load images
hpn -a load --load-mode 2

# Push images
hpn -a push -f images.txt -r registry.example.com --push-mode 2
```

## Configuration

Harpoon can be configured using a YAML configuration file. See [Configuration Guide](configuration.md) for details.

## Advanced Usage

### Batch Operations

Harpoon supports batch operations for efficient image management:

```bash
# Create image list
echo "nginx:latest" > images.txt
echo "alpine:3.18" >> images.txt

# Process multiple images
hpn -a pull -f images.txt
```

### Runtime Selection

Harpoon supports multiple container runtimes:

```bash
# Use specific runtime
hpn --runtime podman -a pull -f images.txt

# Auto-fallback mode
hpn --auto-fallback -a pull -f images.txt
```

## Troubleshooting

For common issues and solutions, see the [Troubleshooting Guide](troubleshooting.md).

## Examples

For more examples, see the [Examples](examples.md) documentation.
EOF
        print_success "用户指南模板已创建: $user_doc"
    else
        # 更新现有用户指南中的示例
        print_info "用户指南已存在，检查是否需要更新示例..."
        
        # 这里可以添加更复杂的更新逻辑
        # 例如：更新版本号、命令示例等
        
        # 获取当前版本
        CURRENT_VERSION=$(./hpn --version 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "v1.1.0")
        
        # 更新版本引用
        if grep -q "version" "$user_doc"; then
            sed -i.bak "s/v[0-9]\+\.[0-9]\+\.[0-9]\+/$CURRENT_VERSION/g" "$user_doc"
            rm -f "$user_doc.bak"
            print_success "用户指南中的版本信息已更新"
        fi
    fi
}

# 更新变更日志
update_changelog() {
    print_info "更新变更日志..."
    
    if [ "$DRY_RUN" = true ]; then
        print_info "预览模式: 将更新变更日志"
        return 0
    fi
    
    # 使用现有的changelog更新脚本
    if [ -f "scripts/update-changelog.sh" ]; then
        chmod +x scripts/update-changelog.sh
        ./scripts/update-changelog.sh -f "$OUTPUT_DIR/changelog.md"
        print_success "变更日志已更新"
    else
        print_warning "changelog更新脚本不存在，跳过"
    fi
}

# 验证文档质量
validate_documentation() {
    print_info "验证文档质量..."
    
    local issues=0
    
    # 检查必需文件
    required_files=(
        "$OUTPUT_DIR/api-reference.md"
        "$OUTPUT_DIR/cli-reference.md"
        "$OUTPUT_DIR/user-guide.md"
        "$OUTPUT_DIR/changelog.md"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            print_warning "缺少文件: $file"
            ((issues++))
        fi
    done
    
    # 检查文档内容
    if [ -f "$OUTPUT_DIR/user-guide.md" ]; then
        if ! grep -q "## Getting Started" "$OUTPUT_DIR/user-guide.md"; then
            print_warning "用户指南缺少 'Getting Started' 部分"
            ((issues++))
        fi
    fi
    
    if [ -f "$OUTPUT_DIR/api-reference.md" ]; then
        if ! grep -q "## Packages" "$OUTPUT_DIR/api-reference.md"; then
            print_warning "API文档缺少包信息"
            ((issues++))
        fi
    fi
    
    if [ $issues -eq 0 ]; then
        print_success "文档质量验证通过"
    else
        print_warning "发现 $issues 个文档质量问题"
    fi
    
    return $issues
}

# 生成文档索引
generate_index() {
    print_info "生成文档索引..."
    
    local index_file="$OUTPUT_DIR/README.md"
    
    if [ "$DRY_RUN" = true ]; then
        print_info "预览模式: 将生成 $index_file"
        return 0
    fi
    
    cat > "$index_file" << 'EOF'
# Harpoon Documentation

Welcome to the Harpoon documentation! This directory contains comprehensive documentation for the Harpoon container image management tool.

## Documentation Structure

### Getting Started
- [Installation Guide](installation.md) - How to install Harpoon
- [Quick Start Guide](quickstart.md) - Get up and running quickly
- [Configuration Guide](configuration.md) - Configuration options and examples

### User Documentation
- [User Guide](user-guide.md) - Comprehensive usage guide
- [Examples](examples.md) - Real-world usage examples
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

### Reference Documentation
- [API Reference](api-reference.md) - Detailed API documentation
- [CLI Reference](cli-reference.md) - Command-line interface reference
- [Architecture](architecture.md) - System architecture and design

### Development
- [Development Guide](development.md) - Contributing and development setup
- [Security Guide](security.md) - Security best practices
- [Changelog](changelog.md) - Version history and changes

## Quick Links

- [GitHub Repository](https://github.com/harpoon/hpn)
- [Issue Tracker](https://github.com/harpoon/hpn/issues)
- [Discussions](https://github.com/harpoon/hpn/discussions)

## Contributing

We welcome contributions to improve the documentation! Please see our [Development Guide](development.md) for information on how to contribute.
EOF
    
    print_success "文档索引已生成: $index_file"
}

# 主函数
main() {
    print_info "文档更新脚本启动"
    
    validate_parameters
    check_prerequisites
    
    case "$UPDATE_TYPE" in
        "api")
            update_api_docs
            ;;
        "cli")
            update_cli_docs
            ;;
        "user")
            update_user_guide
            ;;
        "changelog")
            update_changelog
            ;;
        "all")
            update_api_docs
            update_cli_docs
            update_user_guide
            update_changelog
            generate_index
            ;;
    esac
    
    if [ "$DRY_RUN" = false ]; then
        validate_documentation
    fi
    
    print_success "文档更新完成！"
    print_info "文档位置: $OUTPUT_DIR/"
}

# 运行主函数
main "$@"