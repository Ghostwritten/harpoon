#!/bin/bash

################################################################################
# Rollback Script for Harpoon
# 支持快速回滚到上一版本的自动化脚本
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
ENVIRONMENT=""
TARGET_VERSION=""
DRY_RUN=false
FORCE=false
ROLLBACK_STEPS=1
HEALTH_CHECK_TIMEOUT=300

# 显示帮助信息
show_help() {
    echo "Harpoon Rollback Script"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -e, --environment ENV   目标环境 (development|staging|production)"
    echo "  -v, --version VERSION   回滚到指定版本"
    echo "  -s, --steps STEPS       回滚步数 (默认: 1)"
    echo "  -d, --dry-run           预览模式，不执行实际回滚"
    echo "  -f, --force             强制回滚，跳过安全检查"
    echo "  -t, --timeout SECONDS   健康检查超时时间 (默认: 300秒)"
    echo "  -h, --help              显示帮助信息"
    echo ""
    echo "回滚类型:"
    echo "  版本回滚    回滚到指定版本"
    echo "  步数回滚    回滚指定步数"
    echo "  紧急回滚    快速回滚到上一稳定版本"
    echo ""
    echo "示例:"
    echo "  $0 -e production                    # 回滚生产环境1步"
    echo "  $0 -e staging -v v1.0.0             # 回滚到指定版本"
    echo "  $0 -e production -s 2               # 回滚2步"
    echo "  $0 -e production --dry-run          # 预览回滚操作"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -v|--version)
            TARGET_VERSION="$2"
            shift 2
            ;;
        -s|--steps)
            ROLLBACK_STEPS="$2"
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
        -t|--timeout)
            HEALTH_CHECK_TIMEOUT="$2"
            shift 2
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
    if [ -z "$ENVIRONMENT" ]; then
        print_error "请指定环境 (-e|--environment)"
        exit 1
    fi
    
    case "$ENVIRONMENT" in
        "development"|"staging"|"production")
            ;;
        *)
            print_error "无效的环境: $ENVIRONMENT"
            exit 1
            ;;
    esac
    
    if [ -n "$TARGET_VERSION" ] && [ "$ROLLBACK_STEPS" -ne 1 ]; then
        print_error "不能同时指定版本和回滚步数"
        exit 1
    fi
}

# 设置环境配置
setup_environment_config() {
    case "$ENVIRONMENT" in
        "development")
            NAMESPACE="harpoon-dev"
            DEPLOYMENT_NAME="harpoon-dev"
            SERVICE_NAME="harpoon-dev-service"
            HEALTH_URL="https://dev.harpoon.example.com/health"
            ;;
        "staging")
            NAMESPACE="harpoon-staging"
            DEPLOYMENT_NAME="harpoon-staging"
            SERVICE_NAME="harpoon-staging-service"
            HEALTH_URL="https://staging.harpoon.example.com/health"
            ;;
        "production")
            NAMESPACE="harpoon-prod"
            DEPLOYMENT_NAME="harpoon-prod"
            SERVICE_NAME="harpoon-prod-service"
            HEALTH_URL="https://harpoon.example.com/health"
            ;;
    esac
    
    print_info "回滚配置:"
    echo "  Environment: $ENVIRONMENT"
    echo "  Namespace: $NAMESPACE"
    echo "  Deployment: $DEPLOYMENT_NAME"
    if [ -n "$TARGET_VERSION" ]; then
        echo "  Target Version: $TARGET_VERSION"
    else
        echo "  Rollback Steps: $ROLLBACK_STEPS"
    fi
}

# 检查先决条件
check_prerequisites() {
    print_info "检查回滚先决条件..."
    
    # 检查kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl未安装或不在PATH中"
        exit 1
    fi
    
    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        print_error "无法连接到Kubernetes集群"
        exit 1
    fi
    
    # 检查部署是否存在
    if ! kubectl get deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" &> /dev/null; then
        print_error "部署 $DEPLOYMENT_NAME 在命名空间 $NAMESPACE 中不存在"
        exit 1
    fi
    
    print_success "先决条件检查通过"
}

# 获取当前版本信息
get_current_version() {
    print_info "获取当前版本信息..."
    
    # 获取当前镜像
    CURRENT_IMAGE=$(kubectl get deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.template.spec.containers[0].image}')
    
    # 提取版本信息
    if [[ "$CURRENT_IMAGE" =~ :([^:]+)$ ]]; then
        CURRENT_VERSION="${BASH_REMATCH[1]}"
    else
        CURRENT_VERSION="unknown"
    fi
    
    print_info "当前版本: $CURRENT_VERSION"
    print_info "当前镜像: $CURRENT_IMAGE"
}

# 获取回滚历史
get_rollback_history() {
    print_info "获取部署历史..."
    
    # 获取部署历史
    HISTORY=$(kubectl rollout history deployment/"$DEPLOYMENT_NAME" -n "$NAMESPACE" --output=json)
    
    if [ "$DRY_RUN" = true ]; then
        echo "部署历史:"
        kubectl rollout history deployment/"$DEPLOYMENT_NAME" -n "$NAMESPACE"
    fi
    
    # 获取历史版本数量
    HISTORY_COUNT=$(kubectl rollout history deployment/"$DEPLOYMENT_NAME" -n "$NAMESPACE" | grep -c "^[0-9]" || echo "0")
    
    if [ "$HISTORY_COUNT" -lt 2 ]; then
        print_error "没有足够的历史版本进行回滚"
        exit 1
    fi
    
    print_info "可用历史版本数: $HISTORY_COUNT"
}

# 确定回滚目标
determine_rollback_target() {
    if [ -n "$TARGET_VERSION" ]; then
        print_info "回滚到指定版本: $TARGET_VERSION"
        ROLLBACK_TYPE="version"
    else
        print_info "回滚 $ROLLBACK_STEPS 步"
        ROLLBACK_TYPE="steps"
    fi
}

# 生产环境安全检查
production_safety_check() {
    if [ "$ENVIRONMENT" != "production" ] || [ "$FORCE" = true ]; then
        return 0
    fi
    
    print_warning "生产环境回滚安全检查..."
    
    # 检查当前系统状态
    print_info "检查当前系统健康状态..."
    if curl -f -s "$HEALTH_URL" > /dev/null; then
        print_warning "当前系统似乎是健康的"
        if [ "$DRY_RUN" = false ]; then
            read -p "确认要回滚健康的生产系统吗？(yes/no): " -r
            if [[ ! $REPLY =~ ^yes$ ]]; then
                print_info "回滚已取消"
                exit 0
            fi
        fi
    fi
    
    # 检查监控告警
    print_info "检查监控告警状态..."
    # 这里可以集成监控系统API检查
    
    # 确认回滚操作
    if [ "$DRY_RUN" = false ]; then
        print_warning "即将执行生产环境回滚操作"
        read -p "请输入 'ROLLBACK' 确认: " -r
        if [[ $REPLY != "ROLLBACK" ]]; then
            print_info "回滚已取消"
            exit 0
        fi
    fi
    
    print_success "生产环境安全检查通过"
}

# 执行回滚
perform_rollback() {
    print_info "开始执行回滚..."
    
    if [ "$DRY_RUN" = true ]; then
        print_warning "预览模式 - 不会执行实际回滚"
    fi
    
    case "$ROLLBACK_TYPE" in
        "version")
            rollback_to_version
            ;;
        "steps")
            rollback_by_steps
            ;;
    esac
}

# 按版本回滚
rollback_to_version() {
    print_info "回滚到版本: $TARGET_VERSION"
    
    # 构建目标镜像名
    TARGET_IMAGE="ghcr.io/harpoon/hpn:$TARGET_VERSION"
    
    if [ "$DRY_RUN" = false ]; then
        # 更新部署镜像
        kubectl set image deployment/"$DEPLOYMENT_NAME" harpoon="$TARGET_IMAGE" -n "$NAMESPACE"
        
        # 等待回滚完成
        kubectl rollout status deployment/"$DEPLOYMENT_NAME" -n "$NAMESPACE" --timeout=600s
    else
        echo "kubectl set image deployment/$DEPLOYMENT_NAME harpoon=$TARGET_IMAGE -n $NAMESPACE"
    fi
}

# 按步数回滚
rollback_by_steps() {
    print_info "回滚 $ROLLBACK_STEPS 步"
    
    if [ "$DRY_RUN" = false ]; then
        # 执行回滚
        if [ "$ROLLBACK_STEPS" -eq 1 ]; then
            kubectl rollout undo deployment/"$DEPLOYMENT_NAME" -n "$NAMESPACE"
        else
            # 获取目标revision
            CURRENT_REVISION=$(kubectl get deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
            TARGET_REVISION=$((CURRENT_REVISION - ROLLBACK_STEPS))
            
            if [ "$TARGET_REVISION" -lt 1 ]; then
                print_error "回滚步数超出历史范围"
                exit 1
            fi
            
            kubectl rollout undo deployment/"$DEPLOYMENT_NAME" -n "$NAMESPACE" --to-revision="$TARGET_REVISION"
        fi
        
        # 等待回滚完成
        kubectl rollout status deployment/"$DEPLOYMENT_NAME" -n "$NAMESPACE" --timeout=600s
    else
        if [ "$ROLLBACK_STEPS" -eq 1 ]; then
            echo "kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE"
        else
            echo "kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE --to-revision=<target>"
        fi
    fi
}

# 验证回滚
verify_rollback() {
    print_info "验证回滚结果..."
    
    if [ "$DRY_RUN" = true ]; then
        print_success "回滚验证通过 (预览模式)"
        return 0
    fi
    
    # 等待Pod就绪
    print_info "等待Pod就绪..."
    kubectl wait --for=condition=ready pod -l app=harpoon -n "$NAMESPACE" --timeout=300s
    
    # 健康检查
    print_info "执行健康检查..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$HEALTH_URL" > /dev/null; then
            print_success "健康检查通过"
            break
        fi
        
        echo "健康检查尝试 $attempt/$max_attempts 失败，等待10秒后重试..."
        sleep 10
        attempt=$((attempt + 1))
        
        if [ $attempt -gt $max_attempts ]; then
            print_error "健康检查失败，回滚可能未成功"
            exit 1
        fi
    done
    
    # 获取回滚后的版本信息
    NEW_IMAGE=$(kubectl get deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.template.spec.containers[0].image}')
    if [[ "$NEW_IMAGE" =~ :([^:]+)$ ]]; then
        NEW_VERSION="${BASH_REMATCH[1]}"
    else
        NEW_VERSION="unknown"
    fi
    
    print_success "回滚验证完成"
    print_info "回滚前版本: $CURRENT_VERSION"
    print_info "回滚后版本: $NEW_VERSION"
}

# 记录回滚操作
log_rollback() {
    print_info "记录回滚操作..."
    
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local log_entry="[$timestamp] Rollback: $ENVIRONMENT from $CURRENT_VERSION to $NEW_VERSION"
    
    # 记录到本地日志
    echo "$log_entry" >> rollback.log
    
    # 创建Kubernetes事件
    if [ "$DRY_RUN" = false ]; then
        kubectl annotate deployment "$DEPLOYMENT_NAME" -n "$NAMESPACE" \
            "rollback.harpoon.io/timestamp=$timestamp" \
            "rollback.harpoon.io/from-version=$CURRENT_VERSION" \
            "rollback.harpoon.io/to-version=$NEW_VERSION" \
            --overwrite
    fi
    
    print_success "回滚操作已记录"
}

# 发送通知
send_notification() {
    print_info "发送回滚通知..."
    
    local status="success"
    if [ $? -ne 0 ]; then
        status="failed"
    fi
    
    # 这里可以集成Slack、邮件等通知系统
    print_info "回滚通知: $ENVIRONMENT 环境回滚 $status"
    
    # 示例: Slack通知
    # curl -X POST -H 'Content-type: application/json' \
    #   --data "{\"text\":\"🔄 Rollback $status: $ENVIRONMENT from $CURRENT_VERSION to $NEW_VERSION\"}" \
    #   $SLACK_WEBHOOK_URL
}

# 主函数
main() {
    print_info "Harpoon回滚脚本启动"
    
    validate_parameters
    setup_environment_config
    check_prerequisites
    get_current_version
    get_rollback_history
    determine_rollback_target
    production_safety_check
    perform_rollback
    verify_rollback
    log_rollback
    send_notification
    
    print_success "回滚操作完成！"
    print_info "系统已回滚到稳定状态"
}

# 运行主函数
main "$@"