#!/bin/bash

# 生成测试用自签名证书脚本
# 用于测试Kun Gateway的多域名HTTPS功能

set -e

# 配置
CERT_DIR="certs"
DOMAINS=("example.com" "api.example.com" "admin.example.com")

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 创建证书目录
create_cert_dir() {
    if [ ! -d "$CERT_DIR" ]; then
        log_info "创建证书目录: $CERT_DIR"
        mkdir -p "$CERT_DIR"
    fi
}

# 生成自签名证书
generate_cert() {
    local domain=$1
    local cert_file="$CERT_DIR/$domain.crt"
    local key_file="$CERT_DIR/$domain.key"
    
    log_info "生成 $domain 的证书..."
    
    # 生成私钥
    openssl genrsa -out "$key_file" 2048
    
    # 生成证书签名请求
    openssl req -new -key "$key_file" -out "$CERT_DIR/$domain.csr" -subj "/CN=$domain/O=Test/C=CN"
    
    # 生成自签名证书
    openssl x509 -req -in "$CERT_DIR/$domain.csr" -signkey "$key_file" -out "$cert_file" -days 365
    
    # 清理CSR文件
    rm "$CERT_DIR/$domain.csr"
    
    log_info "$domain 证书生成完成: $cert_file, $key_file"
}

# 生成通配符证书
generate_wildcard_cert() {
    local domain="*.example.com"
    local cert_file="$CERT_DIR/wildcard.example.com.crt"
    local key_file="$CERT_DIR/wildcard.example.com.key"
    
    log_info "生成通配符证书: $domain"
    
    # 生成私钥
    openssl genrsa -out "$key_file" 2048
    
    # 生成证书签名请求
    openssl req -new -key "$key_file" -out "$CERT_DIR/wildcard.example.com.csr" -subj "/CN=$domain/O=Test/C=CN"
    
    # 生成自签名证书
    openssl x509 -req -in "$CERT_DIR/wildcard.example.com.csr" -signkey "$key_file" -out "$cert_file" -days 365
    
    # 清理CSR文件
    rm "$CERT_DIR/wildcard.example.com.csr"
    
    log_info "通配符证书生成完成: $cert_file, $key_file"
}

# 验证证书
verify_cert() {
    local domain=$1
    local cert_file="$CERT_DIR/$domain.crt"
    
    log_info "验证 $domain 证书..."
    
    if openssl x509 -in "$cert_file" -text -noout >/dev/null 2>&1; then
        log_info "$domain 证书验证成功"
        
        # 显示证书信息
        echo "证书信息:"
        openssl x509 -in "$cert_file" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After)"
        echo ""
    else
        log_warn "$domain 证书验证失败"
    fi
}

# 设置证书权限
set_permissions() {
    log_info "设置证书文件权限..."
    chmod 600 "$CERT_DIR"/*.key
    chmod 644 "$CERT_DIR"/*.crt
}

# 主函数
main() {
    log_info "开始生成测试证书..."
    
    # 检查openssl是否安装
    if ! command -v openssl >/dev/null 2>&1; then
        log_warn "openssl未安装，请先安装openssl"
        exit 1
    fi
    
    # 创建证书目录
    create_cert_dir
    
    # 生成各个域名的证书
    for domain in "${DOMAINS[@]}"; do
        generate_cert "$domain"
        verify_cert "$domain"
    done
    
    # 生成通配符证书
    generate_wildcard_cert
    verify_cert "wildcard.example.com"
    
    # 设置权限
    set_permissions
    
    log_info "所有测试证书生成完成！"
    log_info "证书文件位置: $CERT_DIR/"
    
    echo ""
    echo "生成的证书文件:"
    ls -la "$CERT_DIR"/
}

# 帮助信息
show_help() {
    echo "测试证书生成脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -d, --dir DIR  指定证书目录 (默认: certs)"
    echo ""
    echo "示例:"
    echo "  $0"
    echo "  $0 -d /tmp/certs"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--dir)
            CERT_DIR="$2"
            shift 2
            ;;
        *)
            log_warn "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 运行主函数
main "$@" 