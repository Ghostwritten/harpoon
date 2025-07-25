#!/bin/bash

################################################################################
# Changelog Update Script
# 自动更新changelog文件
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
CHANGELOG_FILE="docs/changelog.md"
VERSION=""
SINCE_TAG=""
OUTPUT_FORMAT="markdown"
DRY_RUN=false

# 显示帮助信息
show_help() {
    echo "Changelog Update Script"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -v, --version VERSION   指定版本号"
    echo "  -s, --since TAG         从指定标签开始生成"
    echo "  -f, --file FILE         changelog文件路径 (默认: docs/changelog.md)"
    echo "  -o, --format FORMAT     输出格式 (markdown|json) (默认: markdown)"
    echo "  -d, --dry-run           只显示内容，不写入文件"
    echo "  -h, --help              显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -v v1.1.0                    # 为v1.1.0生成changelog"
    echo "  $0 -s v1.0.0 -v v1.1.0          # 从v1.0.0到v1.1.0的变更"
    echo "  $0 -d                           # 预览模式"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -s|--since)
            SINCE_TAG="$2"
            shift 2
            ;;
        -f|--file)
            CHANGELOG_FILE="$2"
            shift 2
            ;;
        -o|--format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        -d|--dry-run)
            DRY_RUN=true
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

# 检查Git仓库
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "当前目录不是Git仓库"
    exit 1
fi

# 获取版本信息
get_version_info() {
    if [ -z "$VERSION" ]; then
        # 尝试获取最新标签
        if git describe --tags --abbrev=0 2>/dev/null; then
            VERSION=$(git describe --tags --abbrev=0 2>/dev/null)
        else
            VERSION="Unreleased"
        fi
    fi
    
    if [ -z "$SINCE_TAG" ]; then
        # 获取上一个标签
        if [ "$VERSION" != "Unreleased" ]; then
            SINCE_TAG=$(git describe --tags --abbrev=0 "$VERSION^" 2>/dev/null || echo "")
        fi
    fi
    
    print_info "版本: $VERSION"
    print_info "起始标签: ${SINCE_TAG:-"(从开始)"}"
}

# 分析提交类型
analyze_commit_type() {
    local message="$1"
    
    if [[ "$message" =~ ^feat(\(.+\))?: ]]; then
        echo "feature"
    elif [[ "$message" =~ ^fix(\(.+\))?: ]]; then
        echo "bugfix"
    elif [[ "$message" =~ ^docs(\(.+\))?: ]]; then
        echo "documentation"
    elif [[ "$message" =~ ^style(\(.+\))?: ]]; then
        echo "style"
    elif [[ "$message" =~ ^refactor(\(.+\))?: ]]; then
        echo "refactor"
    elif [[ "$message" =~ ^test(\(.+\))?: ]]; then
        echo "test"
    elif [[ "$message" =~ ^chore(\(.+\))?: ]]; then
        echo "chore"
    elif [[ "$message" =~ ^perf(\(.+\))?: ]]; then
        echo "performance"
    elif [[ "$message" =~ ^ci(\(.+\))?: ]]; then
        echo "ci"
    elif [[ "$message" =~ ^build(\(.+\))?: ]]; then
        echo "build"
    elif [[ "$message" =~ ^revert(\(.+\))?: ]]; then
        echo "revert"
    else
        echo "other"
    fi
}

# 提取提交范围
extract_scope() {
    local message="$1"
    
    if [[ "$message" =~ ^\w+\(([^)]+)\): ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        echo ""
    fi
}

# 检查是否为破坏性变更
is_breaking_change() {
    local message="$1"
    
    if [[ "$message" =~ BREAKING[[:space:]]CHANGE ]] || [[ "$message" =~ ! ]]; then
        return 0
    else
        return 1
    fi
}

# 生成变更日志
generate_changelog() {
    local range=""
    
    if [ -n "$SINCE_TAG" ]; then
        range="${SINCE_TAG}..HEAD"
    else
        range="HEAD"
    fi
    
    print_info "分析提交范围: $range"
    
    # 获取提交信息
    local commits=$(git log --pretty=format:"%H|%s|%an|%ad" --date=short "$range" 2>/dev/null)
    
    if [ -z "$commits" ]; then
        print_warning "没有找到提交记录"
        return 1
    fi
    
    # 分类提交
    declare -A categories
    categories[feature]=""
    categories[bugfix]=""
    categories[documentation]=""
    categories[performance]=""
    categories[refactor]=""
    categories[test]=""
    categories[ci]=""
    categories[build]=""
    categories[style]=""
    categories[chore]=""
    categories[revert]=""
    categories[other]=""
    categories[breaking]=""
    
    local commit_count=0
    
    while IFS='|' read -r hash subject author date; do
        if [ -z "$hash" ]; then continue; fi
        
        commit_count=$((commit_count + 1))
        
        local type=$(analyze_commit_type "$subject")
        local scope=$(extract_scope "$subject")
        
        # 格式化提交信息
        local formatted_subject="$subject"
        if [ -n "$scope" ]; then
            formatted_subject=$(echo "$subject" | sed "s/^[^:]*: //")
        fi
        
        local commit_line="- $formatted_subject ([${hash:0:7}](../../commit/$hash))"
        
        # 检查破坏性变更
        if is_breaking_change "$subject"; then
            categories[breaking]+="$commit_line"$'\n'
        fi
        
        # 添加到对应分类
        categories[$type]+="$commit_line"$'\n'
        
    done <<< "$commits"
    
    print_info "分析了 $commit_count 个提交"
    
    # 生成markdown格式的changelog
    generate_markdown_changelog
}

# 生成Markdown格式的changelog
generate_markdown_changelog() {
    local output=""
    local date=$(date -u +"%Y-%m-%d")
    
    # 标题
    if [ "$VERSION" = "Unreleased" ]; then
        output+="## [Unreleased]"$'\n'
    else
        output+="## [$VERSION] - $date"$'\n'
    fi
    output+=""$'\n'
    
    # 破坏性变更 (最重要，放在最前面)
    if [ -n "${categories[breaking]}" ]; then
        output+="### 💥 Breaking Changes"$'\n'
        output+=""$'\n'
        output+="${categories[breaking]}"
        output+=""$'\n'
    fi
    
    # 新功能
    if [ -n "${categories[feature]}" ]; then
        output+="### ✨ Features"$'\n'
        output+=""$'\n'
        output+="${categories[feature]}"
        output+=""$'\n'
    fi
    
    # Bug修复
    if [ -n "${categories[bugfix]}" ]; then
        output+="### 🐛 Bug Fixes"$'\n'
        output+=""$'\n'
        output+="${categories[bugfix]}"
        output+=""$'\n'
    fi
    
    # 性能改进
    if [ -n "${categories[performance]}" ]; then
        output+="### ⚡ Performance"$'\n'
        output+=""$'\n'
        output+="${categories[performance]}"
        output+=""$'\n'
    fi
    
    # 重构
    if [ -n "${categories[refactor]}" ]; then
        output+="### ♻️ Refactor"$'\n'
        output+=""$'\n'
        output+="${categories[refactor]}"
        output+=""$'\n'
    fi
    
    # 文档
    if [ -n "${categories[documentation]}" ]; then
        output+="### 📚 Documentation"$'\n'
        output+=""$'\n'
        output+="${categories[documentation]}"
        output+=""$'\n'
    fi
    
    # 测试
    if [ -n "${categories[test]}" ]; then
        output+="### 🧪 Tests"$'\n'
        output+=""$'\n'
        output+="${categories[test]}"
        output+=""$'\n'
    fi
    
    # CI/CD
    if [ -n "${categories[ci]}" ]; then
        output+="### 👷 CI/CD"$'\n'
        output+=""$'\n'
        output+="${categories[ci]}"
        output+=""$'\n'
    fi
    
    # 构建系统
    if [ -n "${categories[build]}" ]; then
        output+="### 🔨 Build System"$'\n'
        output+=""$'\n'
        output+="${categories[build]}"
        output+=""$'\n'
    fi
    
    # 其他变更
    local other_changes=""
    other_changes+="${categories[style]}"
    other_changes+="${categories[chore]}"
    other_changes+="${categories[revert]}"
    other_changes+="${categories[other]}"
    
    if [ -n "$other_changes" ]; then
        output+="### 🔧 Other Changes"$'\n'
        output+=""$'\n'
        output+="$other_changes"
        output+=""$'\n'
    fi
    
    echo "$output"
}

# 更新changelog文件
update_changelog_file() {
    local new_content="$1"
    
    if [ "$DRY_RUN" = true ]; then
        print_info "预览模式 - 生成的changelog内容:"
        echo "----------------------------------------"
        echo "$new_content"
        echo "----------------------------------------"
        return 0
    fi
    
    # 创建目录
    local dir=$(dirname "$CHANGELOG_FILE")
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        print_info "创建目录: $dir"
    fi
    
    # 备份现有文件
    if [ -f "$CHANGELOG_FILE" ]; then
        cp "$CHANGELOG_FILE" "${CHANGELOG_FILE}.bak"
        print_info "备份现有文件: ${CHANGELOG_FILE}.bak"
    fi
    
    # 合并新内容和现有内容
    if [ -f "$CHANGELOG_FILE" ] && [ "$VERSION" != "Unreleased" ]; then
        # 如果是正式版本，插入到现有changelog中
        {
            echo "# Changelog"
            echo ""
            echo "All notable changes to this project will be documented in this file."
            echo ""
            echo "The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),"
            echo "and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)."
            echo ""
            echo "$new_content"
            
            # 添加现有内容（跳过标题部分）
            if grep -q "^## \[" "$CHANGELOG_FILE"; then
                sed -n '/^## \[/,$p' "$CHANGELOG_FILE"
            fi
        } > "${CHANGELOG_FILE}.new"
    else
        # 创建新文件或替换Unreleased部分
        {
            echo "# Changelog"
            echo ""
            echo "All notable changes to this project will be documented in this file."
            echo ""
            echo "The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),"
            echo "and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)."
            echo ""
            echo "$new_content"
        } > "${CHANGELOG_FILE}.new"
    fi
    
    # 替换原文件
    mv "${CHANGELOG_FILE}.new" "$CHANGELOG_FILE"
    
    print_success "Changelog已更新: $CHANGELOG_FILE"
}

# 主函数
main() {
    get_version_info
    
    local changelog_content=$(generate_changelog)
    if [ $? -ne 0 ]; then
        print_error "生成changelog失败"
        exit 1
    fi
    
    update_changelog_file "$changelog_content"
    
    if [ "$DRY_RUN" = false ]; then
        print_success "Changelog更新完成"
        print_info "文件位置: $CHANGELOG_FILE"
        
        if [ -f "${CHANGELOG_FILE}.bak" ]; then
            print_info "备份文件: ${CHANGELOG_FILE}.bak"
        fi
    fi
}

# 运行主函数
main "$@"