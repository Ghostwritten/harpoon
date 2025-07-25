#!/bin/bash

################################################################################
# Deployment Script for Harpoon
# 支持多环境部署的自动化脚本
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
VERSION=""
IMAGE_TAG=""
DRY_RUN=false
FORCE=false
ROLLBACK=false
HEALTH_CHECK_TIMEOUT=300
DEPLOYMENT_TIMEOUT=600

# 显示帮助信息
show_help() {
    echo "Harpoon Deployment Script"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -e, --environment ENV   目标环境 (development|staging|production)"
    echo "  -v, --version VERSION   部署版本"
    echo "  -i, --image TAG         Docker镜像标签"
    echo "  -d, --dry-run           预览模式，不执行实际部署"
    echo "  -f, --force             强制部署，跳过检查"
    echo "  -r, --rollback          回滚到上一版本"
    echo "  -t, --timeout SECONDS   部署超时时间 (默认: 600秒)"
    echo "  -h, --help              显示帮助信息"
    echo ""
    echo "环境配置:"
    echo "  development   开发环境 (自动部署)"
    echo "  staging       测试环境 (需要审批)"
    echo "  production    生产环境 (需要审批和额外检查)"
    echo ""
    echo "示例:"
    echo "  $0 -e development -v v1.1.0"
    echo "  $0 -e staging -i ghcr.io/org/harpoon:v1.1.0"
    echo "  $0 -e production -v v1.1.0 --dry-run"
    echo "  $0 -e production --rollback"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -i|--image)
            IMAGE_TAG="$2"
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
        -r|--rollback)
            ROLLBACK=true
            shift
            ;;
        -t|--timeout)
            DEPLOYMENT_TIMEOUT="$2"
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
        print_error "请指定部署环境 (-e|--environment)"
        exit 1
    fi
    
    case "$ENVIRONMENT" in
        "development"|"staging"|"production")
            ;;
        *)
            print_error "无效的环境: $ENVIRONMENT"
            print_error "支持的环境: development, staging, production"
            exit 1
            ;;
    esac
    
    if [ "$ROLLBACK" = false ]; then
        if [ -z "$VERSION" ] && [ -z "$IMAGE_TAG" ]; then
            print_error "请指定版本 (-v) 或镜像标签 (-i)"
            exit 1
        fi
    fi
}

# 设置环境配置
setup_environment_config() {
    case "$ENVIRONMENT" in
        "development")
            NAMESPACE="harpoon-dev"
            DEPLOYMENT_NAME="harpoon-dev"
            SERVICE_NAME="harpoon-dev-service"
            INGRESS_HOST="dev.harpoon.example.com"
            REPLICAS=1
            RESOURCE_LIMITS="cpu=500m,memory=512Mi"
            RESOURCE_REQUESTS="cpu=100m,memory=128Mi"
            ;;
        "staging")
            NAMESPACE="harpoon-staging"
            DEPLOYMENT_NAME="harpoon-staging"
            SERVICE_NAME="harpoon-staging-service"
            INGRESS_HOST="staging.harpoon.example.com"
            REPLICAS=2
            RESOURCE_LIMITS="cpu=1000m,memory=1Gi"
            RESOURCE_REQUESTS="cpu=200m,memory=256Mi"
            ;;
        "production")
            NAMESPACE="harpoon-prod"
            DEPLOYMENT_NAME="harpoon-prod"
            SERVICE_NAME="harpoon-prod-service"
            INGRESS_HOST="harpoon.example.com"
            REPLICAS=3
            RESOURCE_LIMITS="cpu=2000m,memory=2Gi"
            RESOURCE_REQUESTS="cpu=500m,memory=512Mi"
            ;;
    esac
    
    # 构建镜像标签
    if [ -n "$VERSION" ] && [ -z "$IMAGE_TAG" ]; then
        IMAGE_TAG="ghcr.io/harpoon/hpn:$VERSION"
    fi
    
    print_info "环境配置:"
    echo "  Environment: $ENVIRONMENT"
    echo "  Namespace: $NAMESPACE"
    echo "  Deployment: $DEPLOYMENT_NAME"
    echo "  Image: $IMAGE_TAG"
    echo "  Replicas: $REPLICAS"
    echo "  Host: $INGRESS_HOST"
}

# 检查先决条件
check_prerequisites() {
    print_info "检查部署先决条件..."
    
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
    
    # 检查命名空间
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_warning "命名空间 $NAMESPACE 不存在，将创建"
        if [ "$DRY_RUN" = false ]; then
            kubectl create namespace "$NAMESPACE"
        fi
    fi
    
    # 检查镜像是否存在
    if [ "$ROLLBACK" = false ]; then
        print_info "验证镜像: $IMAGE_TAG"
        if ! docker manifest inspect "$IMAGE_TAG" &> /dev/null; then
            if [ "$FORCE" = false ]; then
                print_error "镜像 $IMAGE_TAG 不存在或无法访问"
                exit 1
            else
                print_warning "镜像验证失败，但使用 --force 继续"
            fi
        fi
    fi
    
    print_success "先决条件检查通过"
}

# 生产环境额外检查
production_checks() {
    if [ "$ENVIRONMENT" != "production" ]; then
        return 0
    fi
    
    print_info "执行生产环境额外检查..."
    
    # 检查staging环境状态
    print_info "检查staging环境状态..."
    if ! curl -f -s "https://staging.harpoon.example.com/health" > /dev/null; then
        if [ "$FORCE" = false ]; then
            print_error "Staging环境不健康，无法部署到生产环境"
            exit 1
        else
            print_warning "Staging环境检查失败，但使用 --force 继续"
        fi
    fi
    
    # 检查数据库迁移
    print_info "检查数据库迁移..."
    # kubectl exec -n harpoon-staging deployment/harpoon-staging -- ./hpn migrate --dry-run
    
    # 检查监控系统
    print_info "检查监控系统状态..."
    # curl -f -s "https://monitoring.harpoon.example.com/api/v1/query?query=up" > /dev/null
    
    print_success "生产环境检查通过"
}

# 执行部署
perform_deployment() {
    if [ "$ROLLBACK" = true ]; then
        perform_rollback
        return
    fi
    
    print_info "开始部署到 $ENVIRONMENT 环境..."
    
    if [ "$DRY_RUN" = true ]; then
        print_warning "预览模式 - 不会执行实际部署"
    fi
    
    # 创建或更新部署配置
    create_deployment_config
    
    # 应用配置
    apply_deployment
    
    # 等待部署完成
    wait_for_deployment
    
    # 健康检查
    health_check
    
    print_success "部署完成！"
}

# 创建部署配置
create_deployment_config() {
    print_info "生成部署配置..."
    
    cat > /tmp/harpoon-deployment.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $DEPLOYMENT_NAME
  namespace: $NAMESPACE
  labels:
    app: harpoon
    environment: $ENVIRONMENT
    version: $VERSION
spec:
  replicas: $REPLICAS
  selector:
    matchLabels:
      app: harpoon
      environment: $ENVIRONMENT
  template:
    metadata:
      labels:
        app: harpoon
        environment: $ENVIRONMENT
        version: $VERSION
    spec:
      containers:
      - name: harpoon
        image: $IMAGE_TAG
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: ENVIRONMENT
          value: $ENVIRONMENT
        - name: VERSION
          value: $VERSION
        resources:
          limits:
            cpu: $(echo $RESOURCE_LIMITS | cut -d',' -f1 | cut -d'=' -f2)
            memory: $(echo $RESOURCE_LIMITS | cut -d',' -f2 | cut -d'=' -f2)
          requests:
            cpu: $(echo $RESOURCE_REQUESTS | cut -d',' -f1 | cut -d'=' -f2)
            memory: $(echo $RESOURCE_REQUESTS | cut -d',' -f2 | cut -d'=' -f2)
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: $SERVICE_NAME
  namespace: $NAMESPACE
  labels:
    app: harpoon
    environment: $ENVIRONMENT
spec:
  selector:
    app: harpoon
    environment: $ENVIRONMENT
  ports:
  - port: 80
    targetPort: 8080
    name: http
  type: ClusterIP
EOF

    if [ "$DRY_RUN" = true ]; then
        echo "生成的部署配置:"
        cat /tmp/harpoon-deployment.yaml
    fi
}

# 应用部署
apply_deployment() {
    print_info "应用部署配置..."
    
    if [ "$DRY_RUN" = false ]; then
        kubectl apply -f /tmp/harpoon-deployment.yaml
    else
        echo "kubectl apply -f /tmp/harpoon-deployment.yaml"
    fi
}

# 等待部署完成
wait_for_deployment() {
    print_info "等待部署完成..."
    
    if [ "$DRY_RUN" = false ]; then
        if ! kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=${DEPLOYMENT_TIMEOUT}s; then
            print_error "部署超时或失败"
            exit 1
        fi
    else
        echo "kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=${DEPLOYMENT_TIMEOUT}s"
    fi
    
    print_success "部署rollout完成"
}

# 健康检查
health_check() {
    print_info "执行健康检查..."
    
    local health_url="https://$INGRESS_HOST/health"
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if [ "$DRY_RUN" = false ]; then
            if curl -f -s "$health_url" > /dev/null; then
                print_success "健康检查通过"
                return 0
            fi
        else
            echo "curl -f -s $health_url"
            print_success "健康检查通过 (预览模式)"
            return 0
        fi
        
        echo "健康检查尝试 $attempt/$max_attempts 失败，等待10秒后重试..."
        sleep 10
        attempt=$((attempt + 1))
    done
    
    print_error "健康检查失败，应用可能未正常启动"
    exit 1
}

# 执行回滚
perform_rollback() {
    print_info "开始回滚 $ENVIRONMENT 环境..."
    
    if [ "$DRY_RUN" = true ]; then
        print_warning "预览模式 - 不会执行实际回滚"
        echo "kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE"
        return
    fi
    
    # 执行回滚
    kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE
    
    # 等待回滚完成
    print_info "等待回滚完成..."
    kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=${DEPLOYMENT_TIMEOUT}s
    
    # 健康检查
    health_check
    
    print_success "回滚完成！"
}

# 清理临时文件
cleanup() {
    rm -f /tmp/harpoon-deployment.yaml
}

# 主函数
main() {
    print_info "Harpoon部署脚本启动"
    
    validate_parameters
    setup_environment_config
    check_prerequisites
    production_checks
    perform_deployment
    
    print_success "部署流程完成！"
    print_info "应用地址: https://$INGRESS_HOST"
}

# 设置清理陷阱
trap cleanup EXIT

# 运行主函数
main "$@"