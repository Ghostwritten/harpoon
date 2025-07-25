#!/bin/bash

################################################################################
# Build Script for Harpoon
# 自动注入版本信息并构建二进制文件
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
BINARY_NAME="hpn"
OUTPUT_DIR="."
PLATFORMS=""
LDFLAGS_EXTRA=""
BUILD_TAGS=""
VERBOSE=false

# 显示帮助信息
show_help() {
    echo "Harpoon Build Script"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -o, --output DIR        输出目录 (默认: .)"
    echo "  -p, --platforms LIST    构建平台列表 (例如: linux/amd64,darwin/amd64)"
    echo "  -t, --tags TAGS         构建标签"
    echo "  -l, --ldflags FLAGS     额外的ldflags"
    echo "  -v, --verbose           详细输出"
    echo "  -h, --help              显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                                    # 构建当前平台"
    echo "  $0 -p linux/amd64,darwin/amd64       # 构建指定平台"
    echo "  $0 -o dist -p all                    # 构建所有平台到dist目录"
    echo "  $0 -t release -v                     # 使用release标签构建"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -p|--platforms)
            PLATFORMS="$2"
            shift 2
            ;;
        -t|--tags)
            BUILD_TAGS="$2"
            shift 2
            ;;
        -l|--ldflags)
            LDFLAGS_EXTRA="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
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

# 检查Go环境
if ! command -v go &> /dev/null; then
    print_error "Go未安装或不在PATH中"
    exit 1
fi

print_info "Go版本: $(go version)"

# 获取版本信息
get_version_info() {
    # 版本号 - 优先使用git tag，否则使用默认值
    if git describe --tags --exact-match HEAD 2>/dev/null; then
        VERSION=$(git describe --tags --exact-match HEAD 2>/dev/null)
    elif git describe --tags --abbrev=0 2>/dev/null; then
        LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null)
        COMMIT_COUNT=$(git rev-list --count ${LAST_TAG}..HEAD 2>/dev/null || echo "0")
        if [ "$COMMIT_COUNT" -gt 0 ]; then
            VERSION="${LAST_TAG}-dev.${COMMIT_COUNT}"
        else
            VERSION="$LAST_TAG"
        fi
    else
        VERSION="v1.1.0-dev"
    fi
    
    # Git信息
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    
    # 构建信息
    BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    BUILD_USER=$(whoami 2>/dev/null || echo "unknown")
    BUILD_HOST=$(hostname 2>/dev/null || echo "unknown")
    
    # Go版本
    GO_VERSION=$(go version | awk '{print $3}')
    
    if [ "$VERBOSE" = true ]; then
        print_info "版本信息:"
        echo "  Version: $VERSION"
        echo "  Git Commit: $GIT_COMMIT"
        echo "  Git Branch: $GIT_BRANCH"
        echo "  Build Date: $BUILD_DATE"
        echo "  Build User: $BUILD_USER"
        echo "  Build Host: $BUILD_HOST"
        echo "  Go Version: $GO_VERSION"
    fi
}

# 构建ldflags
build_ldflags() {
    local ldflags="-s -w"
    
    # 版本信息注入
    ldflags="$ldflags -X github.com/harpoon/hpn/internal/version.Version=$VERSION"
    ldflags="$ldflags -X github.com/harpoon/hpn/internal/version.GitCommit=$GIT_COMMIT"
    ldflags="$ldflags -X github.com/harpoon/hpn/internal/version.GitBranch=$GIT_BRANCH"
    ldflags="$ldflags -X github.com/harpoon/hpn/internal/version.BuildDate=$BUILD_DATE"
    ldflags="$ldflags -X github.com/harpoon/hpn/internal/version.BuildUser=$BUILD_USER"
    ldflags="$ldflags -X github.com/harpoon/hpn/internal/version.BuildHost=$BUILD_HOST"
    
    # 额外的ldflags
    if [ -n "$LDFLAGS_EXTRA" ]; then
        ldflags="$ldflags $LDFLAGS_EXTRA"
    fi
    
    echo "$ldflags"
}

# 构建单个平台
build_platform() {
    local goos=$1
    local goarch=$2
    local output_name=$3
    
    print_info "构建 ${goos}/${goarch}..."
    
    local build_cmd="go build"
    
    # 添加构建标签
    if [ -n "$BUILD_TAGS" ]; then
        build_cmd="$build_cmd -tags '$BUILD_TAGS'"
    fi
    
    # 添加ldflags
    local ldflags=$(build_ldflags)
    build_cmd="$build_cmd -ldflags '$ldflags'"
    
    # 添加输出文件
    build_cmd="$build_cmd -o '$output_name'"
    
    # 添加源码路径
    build_cmd="$build_cmd ./cmd/hpn"
    
    # 设置环境变量并执行构建
    if [ "$VERBOSE" = true ]; then
        echo "执行: GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 $build_cmd"
    fi
    
    if GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 eval $build_cmd; then
        print_success "构建成功: $output_name"
        
        # 显示文件信息
        if [ -f "$output_name" ]; then
            local size=$(ls -lh "$output_name" | awk '{print $5}')
            echo "  文件大小: $size"
            
            # 如果是当前平台，测试版本信息
            if [ "$goos" = "$(go env GOOS)" ] && [ "$goarch" = "$(go env GOARCH)" ]; then
                if [ "$VERBOSE" = true ]; then
                    echo "  测试版本信息:"
                    "$output_name" --version || true
                fi
            fi
        fi
    else
        print_error "构建失败: ${goos}/${goarch}"
        return 1
    fi
}

# 获取所有支持的平台
get_all_platforms() {
    echo "linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"
}

# 主构建函数
main_build() {
    # 获取版本信息
    get_version_info
    
    # 创建输出目录
    if [ ! -d "$OUTPUT_DIR" ]; then
        mkdir -p "$OUTPUT_DIR"
        print_info "创建输出目录: $OUTPUT_DIR"
    fi
    
    # 确定要构建的平台
    local platforms_to_build=""
    
    if [ -z "$PLATFORMS" ]; then
        # 默认构建当前平台
        local current_os=$(go env GOOS)
        local current_arch=$(go env GOARCH)
        platforms_to_build="${current_os}/${current_arch}"
    elif [ "$PLATFORMS" = "all" ]; then
        # 构建所有支持的平台
        platforms_to_build=$(get_all_platforms)
    else
        # 使用指定的平台列表
        platforms_to_build=$(echo "$PLATFORMS" | tr ',' ' ')
    fi
    
    print_info "构建平台: $platforms_to_build"
    
    # 构建每个平台
    local success_count=0
    local total_count=0
    
    for platform in $platforms_to_build; do
        IFS='/' read -r goos goarch <<< "$platform"
        
        # 生成输出文件名
        local output_name="${OUTPUT_DIR}/${BINARY_NAME}-${goos}-${goarch}"
        if [ "$goos" = "windows" ]; then
            output_name="${output_name}.exe"
        fi
        
        # 如果只构建当前平台，使用简单的文件名
        if [ "$platforms_to_build" = "$(go env GOOS)/$(go env GOARCH)" ]; then
            output_name="${OUTPUT_DIR}/${BINARY_NAME}"
            if [ "$goos" = "windows" ]; then
                output_name="${output_name}.exe"
            fi
        fi
        
        total_count=$((total_count + 1))
        
        if build_platform "$goos" "$goarch" "$output_name"; then
            success_count=$((success_count + 1))
        fi
    done
    
    # 构建总结
    echo ""
    print_info "构建总结:"
    echo "  成功: $success_count"
    echo "  失败: $((total_count - success_count))"
    echo "  总计: $total_count"
    
    if [ $success_count -eq $total_count ]; then
        print_success "所有构建都成功完成！"
        return 0
    else
        print_error "有 $((total_count - success_count)) 个构建失败"
        return 1
    fi
}

# 运行主函数
main_build