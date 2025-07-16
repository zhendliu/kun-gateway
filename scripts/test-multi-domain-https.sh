#!/bin/bash

# 多域名HTTPS测试脚本
# 用于测试Kun Gateway的多域名HTTPS功能

set -e

# 配置
GATEWAY_IP="localhost"
GATEWAY_PORT="443"
CONTROL_PLANE_URL="http://localhost:9090"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务状态
check_service() {
    local service_name=$1
    local url=$2
    
    log_info "检查 $service_name 服务状态..."
    if curl -s "$url/api/v1/health" > /dev/null; then
        log_info "$service_name 服务正常"
        return 0
    else
        log_error "$service_name 服务异常"
        return 1
    fi
}

# 上传证书
upload_certificate() {
    local domain=$1
    local cert_file=$2
    local key_file=$3
    
    log_info "上传 $domain 的证书..."
    
    if [ ! -f "$cert_file" ] || [ ! -f "$key_file" ]; then
        log_error "证书文件不存在: $cert_file 或 $key_file"
        return 1
    fi
    
    response=$(curl -s -X POST "$CONTROL_PLANE_URL/api/v1/certificates" \
        -F "domain=$domain" \
        -F "cert_file=@$cert_file" \
        -F "key_file=@$key_file")
    
    if echo "$response" | grep -q '"success":true'; then
        log_info "$domain 证书上传成功"
        return 0
    else
        log_error "$domain 证书上传失败: $response"
        return 1
    fi
}

# 测试HTTPS连接
test_https_connection() {
    local domain=$1
    local expected_status=$2
    
    log_info "测试 $domain 的HTTPS连接..."
    
    # 使用curl测试HTTPS连接
    response=$(curl -s -w "%{http_code}" -H "Host: $domain" \
        "https://$GATEWAY_IP:$GATEWAY_PORT/" -k)
    
    status_code="${response: -3}"
    
    if [ "$status_code" = "$expected_status" ]; then
        log_info "$domain HTTPS连接成功，状态码: $status_code"
        return 0
    else
        log_error "$domain HTTPS连接失败，状态码: $status_code"
        return 1
    fi
}

# 测试SNI功能
test_sni() {
    local domain=$1
    
    log_info "测试 $domain 的SNI功能..."
    
    # 使用openssl测试SNI
    if command -v openssl >/dev/null 2>&1; then
        echo | openssl s_client -connect "$GATEWAY_IP:$GATEWAY_PORT" \
            -servername "$domain" -verify_return_error >/dev/null 2>&1
        
        if [ $? -eq 0 ]; then
            log_info "$domain SNI测试成功"
            return 0
        else
            log_error "$domain SNI测试失败"
            return 1
        fi
    else
        log_warn "openssl未安装，跳过SNI测试"
        return 0
    fi
}

# 查看证书列表
list_certificates() {
    log_info "查看当前证书列表..."
    
    response=$(curl -s "$CONTROL_PLANE_URL/api/v1/certificates")
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
}

# 主函数
main() {
    log_info "开始多域名HTTPS功能测试..."
    
    # 检查服务状态
    if ! check_service "控制面" "$CONTROL_PLANE_URL"; then
        log_error "控制面服务不可用，请先启动服务"
        exit 1
    fi
    
    # 查看当前证书
    list_certificates
    
    # 测试证书上传（如果有证书文件）
    if [ -f "certs/example.com.crt" ] && [ -f "certs/example.com.key" ]; then
        upload_certificate "example.com" "certs/example.com.crt" "certs/example.com.key"
    fi
    
    if [ -f "certs/api.example.com.crt" ] && [ -f "certs/api.example.com.key" ]; then
        upload_certificate "api.example.com" "certs/api.example.com.crt" "certs/api.example.com.key"
    fi
    
    # 查看更新后的证书列表
    list_certificates
    
    # 测试HTTPS连接
    log_info "开始HTTPS连接测试..."
    
    # 测试不同域名的HTTPS连接
    test_https_connection "example.com" "404"  # 404是正常的，因为没有配置路由
    test_https_connection "api.example.com" "404"
    test_https_connection "unknown.com" "404"
    
    # 测试SNI功能
    log_info "开始SNI功能测试..."
    test_sni "example.com"
    test_sni "api.example.com"
    
    log_info "多域名HTTPS功能测试完成！"
}

# 帮助信息
show_help() {
    echo "多域名HTTPS测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -i, --ip IP    指定网关IP地址 (默认: localhost)"
    echo "  -p, --port PORT 指定网关端口 (默认: 443)"
    echo "  -c, --control URL 指定控制面URL (默认: http://localhost:9090)"
    echo ""
    echo "示例:"
    echo "  $0"
    echo "  $0 -i 192.168.1.100 -p 443"
    echo "  $0 -c http://192.168.1.100:9090"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -i|--ip)
            GATEWAY_IP="$2"
            shift 2
            ;;
        -p|--port)
            GATEWAY_PORT="$2"
            shift 2
            ;;
        -c|--control)
            CONTROL_PLANE_URL="$2"
            shift 2
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 运行主函数
main "$@" 