#!/bin/bash

################################################################################
# Version Management Script
# 版本管理辅助脚本
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

# 显示帮助信息
show_help() {
    echo "Version Management Script"
    echo ""
    echo "用法: $0 <command> [options]"
    echo ""
    echo "命令:"
    echo "  current                显示当前版本信息"
    echo "  next <type>            计算下一个版本号"
    echo "  tag <version>          创建版本标签"
    echo "  bump <type>            自动升级版本并创建标签"
    echo "  validate <version>     验证版本号格式"
    echo "  changelog              生成变更日志"
    echo ""
    echo "版本类型 (用于next和bump):"
    echo "  major                  主版本号 (1.0.0 -> 2.0.0)"
    echo "  minor                  次版本号 (1.0.0 -> 1.1.0)"
    echo "  patch                  补丁版本号 (1.0.0 -> 1.0.1)"
    echo "  alpha                  Alpha版本 (1.0.0 -> 1.1.0-alpha.1)"
    echo "  beta                   Beta版本 (1.0.0 -> 1.1.0-beta.1)"
    echo "  rc                     Release Candidate (1.0.0 -> 1.1.0-rc.1)"
    echo ""
    echo "示例:"
    echo "  $0 current"
    echo "  $0 next minor"
    echo "  $0 tag v1.1.0"
    echo "  $0 bump patch"
    echo "  $0 validate v1.2.3"
}

# 获取当前版本信息
get_current_version() {
    # 尝试从git标签获取
    if git describe --tags --exact-match HEAD 2>/dev/null; then
        CURRENT_VERSION=$(git describe --tags --exact-match HEAD 2>/dev/null)
    elif git describe --tags --abbrev=0 2>/dev/null; then
        LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null)
        COMMIT_COUNT=$(git rev-list --count ${LAST_TAG}..HEAD 2>/dev/null || echo "0")
        if [ "$COMMIT_COUNT" -gt 0 ]; then
            CURRENT_VERSION="${LAST_TAG}-dev.${COMMIT_COUNT}"
        else
            CURRENT_VERSION="$LAST_TAG"
        fi
    else
        # 从version.go文件获取
        if [ -f "internal/version/version.go" ]; then
            CURRENT_VERSION=$(grep 'Version.*=' internal/version/version.go | sed 's/.*"\(.*\)".*/\1/')
        else
            CURRENT_VERSION="v1.1.0-dev"
        fi
    fi
    
    echo "$CURRENT_VERSION"
}

# 显示当前版本信息
show_current_version() {
    local version=$(get_current_version)
    local commit=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    local date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    print_info "当前版本信息:"
    echo "  Version: $version"
    echo "  Git Commit: ${commit:0:7}"
    echo "  Git Branch: $branch"
    echo "  Current Date: $date"
    
    # 检查是否有未提交的更改
    if ! git diff --quiet 2>/dev/null; then
        print_warning "工作目录有未提交的更改"
    fi
    
    # 检查是否有未推送的提交
    if git status --porcelain=v1 2>/dev/null | grep -q '^??'; then
        print_warning "有未跟踪的文件"
    fi
}

# 解析版本号
parse_version() {
    local version=$1
    
    # 移除v前缀
    version=${version#v}
    
    # 分离主版本号和预发布标识
    if [[ "$version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)(-([a-z]+)\.?([0-9]+)?)?$ ]]; then
        MAJOR=${BASH_REMATCH[1]}
        MINOR=${BASH_REMATCH[2]}
        PATCH=${BASH_REMATCH[3]}
        PRERELEASE=${BASH_REMATCH[5]}
        PRERELEASE_NUM=${BASH_REMATCH[6]:-1}
    else
        return 1
    fi
    
    return 0
}

# 计算下一个版本号
calculate_next_version() {
    local current_version=$(get_current_version)
    local bump_type=$1
    
    # 移除dev后缀
    current_version=${current_version%-dev*}
    
    if ! parse_version "$current_version"; then
        print_error "无法解析当前版本: $current_version"
        return 1
    fi
    
    case "$bump_type" in
        "major")
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            PRERELEASE=""
            ;;
        "minor")
            MINOR=$((MINOR + 1))
            PATCH=0
            PRERELEASE=""
            ;;
        "patch")
            PATCH=$((PATCH + 1))
            PRERELEASE=""
            ;;
        "alpha")
            if [ -z "$PRERELEASE" ]; then
                MINOR=$((MINOR + 1))
                PATCH=0
                PRERELEASE="alpha"
                PRERELEASE_NUM=1
            elif [ "$PRERELEASE" = "alpha" ]; then
                PRERELEASE_NUM=$((PRERELEASE_NUM + 1))
            else
                print_error "无法从 $PRERELEASE 升级到 alpha"
                return 1
            fi
            ;;
        "beta")
            if [ -z "$PRERELEASE" ] || [ "$PRERELEASE" = "alpha" ]; then
                if [ -z "$PRERELEASE" ]; then
                    MINOR=$((MINOR + 1))
                    PATCH=0
                fi
                PRERELEASE="beta"
                PRERELEASE_NUM=1
            elif [ "$PRERELEASE" = "beta" ]; then
                PRERELEASE_NUM=$((PRERELEASE_NUM + 1))
            else
                print_error "无法从 $PRERELEASE 升级到 beta"
                return 1
            fi
            ;;
        "rc")
            if [ -z "$PRERELEASE" ] || [ "$PRERELEASE" = "alpha" ] || [ "$PRERELEASE" = "beta" ]; then
                if [ -z "$PRERELEASE" ]; then
                    MINOR=$((MINOR + 1))
                    PATCH=0
                fi
                PRERELEASE="rc"
                PRERELEASE_NUM=1
            elif [ "$PRERELEASE" = "rc" ]; then
                PRERELEASE_NUM=$((PRERELEASE_NUM + 1))
            else
                print_error "无法从 $PRERELEASE 升级到 rc"
                return 1
            fi
            ;;
        *)
            print_error "未知的版本类型: $bump_type"
            return 1
            ;;
    esac
    
    # 构建新版本号
    local new_version="v${MAJOR}.${MINOR}.${PATCH}"
    if [ -n "$PRERELEASE" ]; then
        new_version="${new_version}-${PRERELEASE}.${PRERELEASE_NUM}"
    fi
    
    echo "$new_version"
}

# 验证版本号格式
validate_version() {
    local version=$1
    
    if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-z]+\.[0-9]+)?$ ]]; then
        print_error "版本号格式不正确: $version"
        echo "正确格式: v1.2.3 或 v1.2.3-alpha.1"
        return 1
    fi
    
    print_success "版本号格式正确: $version"
    return 0
}

# 创建版本标签
create_tag() {
    local version=$1
    
    if [ -z "$version" ]; then
        print_error "请提供版本号"
        return 1
    fi
    
    # 验证版本号格式
    if ! validate_version "$version"; then
        return 1
    fi
    
    # 检查标签是否已存在
    if git tag -l | grep -q "^${version}$"; then
        print_error "标签 $version 已存在"
        return 1
    fi
    
    # 检查工作目录是否干净
    if ! git diff --quiet || ! git diff --cached --quiet; then
        print_error "工作目录有未提交的更改，请先提交"
        return 1
    fi
    
    # 更新version.go文件
    if [ -f "internal/version/version.go" ]; then
        print_info "更新 internal/version/version.go"
        sed -i.bak "s/Version.*=.*/Version = \"$version\"/" internal/version/version.go
        rm -f internal/version/version.go.bak
        
        # 提交版本文件更改
        git add internal/version/version.go
        git commit -m "chore: bump version to $version"
    fi
    
    # 创建标签
    print_info "创建标签: $version"
    git tag -a "$version" -m "Release $version"
    
    print_success "标签 $version 创建成功"
    print_info "推送标签: git push origin $version"
}

# 自动升级版本
bump_version() {
    local bump_type=$1
    
    if [ -z "$bump_type" ]; then
        print_error "请提供版本类型"
        show_help
        return 1
    fi
    
    local new_version=$(calculate_next_version "$bump_type")
    if [ $? -ne 0 ]; then
        return 1
    fi
    
    print_info "当前版本: $(get_current_version)"
    print_info "新版本: $new_version"
    
    # 确认升级
    read -p "确认升级到 $new_version? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "取消升级"
        return 0
    fi
    
    # 创建标签
    create_tag "$new_version"
}

# 生成变更日志
generate_changelog() {
    local last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    local current_commit=$(git rev-parse HEAD)
    
    print_info "生成变更日志"
    
    if [ -n "$last_tag" ]; then
        echo "## 变更日志 (${last_tag}..HEAD)"
        echo ""
        
        # 获取提交记录
        git log --pretty=format:"- %s (%h)" "${last_tag}..${current_commit}" | while read line; do
            echo "$line"
        done
    else
        echo "## 变更日志 (所有提交)"
        echo ""
        
        git log --pretty=format:"- %s (%h)" | while read line; do
            echo "$line"
        done
    fi
}

# 主函数
main() {
    case "${1:-}" in
        "current")
            show_current_version
            ;;
        "next")
            if [ -z "$2" ]; then
                print_error "请提供版本类型"
                show_help
                exit 1
            fi
            next_version=$(calculate_next_version "$2")
            if [ $? -eq 0 ]; then
                echo "$next_version"
            fi
            ;;
        "tag")
            create_tag "$2"
            ;;
        "bump")
            bump_version "$2"
            ;;
        "validate")
            validate_version "$2"
            ;;
        "changelog")
            generate_changelog
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        "")
            print_error "请指定命令"
            show_help
            exit 1
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"