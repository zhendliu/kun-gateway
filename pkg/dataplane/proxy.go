package dataplane

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

// Proxy 代理服务器
type Proxy struct {
	router    *Router
	log       *logrus.Logger
	metrics   *Metrics
	client    *fasthttp.Client
	ctx       context.Context
	cancel    context.CancelFunc
	connCount int64
	// HTTPS相关配置
	tlsConfig   *tls.Config
	certManager *CertManager
	// HTTP服务器
	httpServer  *http.Server
	httpsServer *http.Server
}

// CertManager 证书管理器
type CertManager struct {
	certs map[string]*tls.Certificate // key: domain
	mu    sync.RWMutex
}

// NewCertManager 创建证书管理器
func NewCertManager() *CertManager {
	return &CertManager{
		certs: make(map[string]*tls.Certificate),
	}
}

// AddCertificate 添加证书
func (cm *CertManager) AddCertificate(domain, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("加载证书失败: %v", err)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.certs[domain] = &cert
	return nil
}

// GetCertificate 获取证书
func (cm *CertManager) GetCertificate(domain string) *tls.Certificate {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.certs[domain]
}

// RemoveCertificate 移除证书
func (cm *CertManager) RemoveCertificate(domain string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.certs, domain)
}

// ListCertificates 列出所有证书域名
func (cm *CertManager) ListCertificates() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	domains := make([]string, 0, len(cm.certs))
	for domain := range cm.certs {
		domains = append(domains, domain)
	}
	return domains
}

// NewProxy 创建代理服务器
func NewProxy(router *Router, log *logrus.Logger) *Proxy {
	ctx, cancel := context.WithCancel(context.Background())

	client := &fasthttp.Client{
		MaxConnsPerHost:     1000,
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        30 * time.Second,
		MaxIdleConnDuration: 10 * time.Second,
	}

	return &Proxy{
		router:      router,
		log:         log,
		metrics:     NewMetrics(),
		client:      client,
		ctx:         ctx,
		cancel:      cancel,
		certManager: NewCertManager(),
	}
}

// Start 启动HTTP代理服务器
func (proxy *Proxy) Start(addr string) error {
	proxy.log.Infof("启动HTTP代理服务器，监听地址: %s", addr)

	// 使用fasthttp启动HTTP服务器
	return fasthttp.ListenAndServe(addr, proxy.handleRequest)
}

// StartTLS 启动HTTPS代理服务器
func (proxy *Proxy) StartTLS(addr string) error {
	proxy.log.Infof("启动HTTPS代理服务器，监听地址: %s", addr)

	// 创建TLS配置，支持SNI
	proxy.tlsConfig = &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			domain := info.ServerName
			if domain == "" {
				// 如果没有SNI，尝试从第一个证书开始
				domains := proxy.certManager.ListCertificates()
				if len(domains) > 0 {
					domain = domains[0]
				}
			}

			cert := proxy.certManager.GetCertificate(domain)
			if cert == nil {
				proxy.log.Warnf("未找到域名 %s 的证书", info.ServerName)
				// 返回第一个可用证书作为默认证书
				domains := proxy.certManager.ListCertificates()
				if len(domains) > 0 {
					cert = proxy.certManager.GetCertificate(domains[0])
				}
			}

			if cert == nil {
				return nil, fmt.Errorf("没有可用的证书")
			}

			proxy.log.Debugf("为域名 %s 选择证书", domain)
			return cert, nil
		},
		MinVersion: tls.VersionTLS12,
	}

	// 创建HTTP服务器
	proxy.httpsServer = &http.Server{
		Addr:      addr,
		Handler:   http.HandlerFunc(proxy.handleHTTPRequest),
		TLSConfig: proxy.tlsConfig,
	}

	// 启动HTTPS服务器
	return proxy.httpsServer.ListenAndServeTLS("", "")
}

// handleHTTPRequest 处理HTTP请求（用于HTTPS）
func (proxy *Proxy) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	atomic.AddInt64(&proxy.connCount, 1)
	defer atomic.AddInt64(&proxy.connCount, -1)

	// 记录请求开始
	proxy.metrics.IncRequests()

	// 获取域名
	domain := r.Host
	if r.TLS != nil && r.TLS.ServerName != "" {
		domain = r.TLS.ServerName
	}

	proxy.log.Debugf("处理HTTPS请求: %s %s", domain, r.URL.Path)

	// 查找路由规则
	rule := proxy.findRouteForDomain(domain, r.URL.Path)
	if rule == nil {
		proxy.log.Warnf("未找到匹配的路由规则: %s%s", domain, r.URL.Path)
		http.Error(w, "404 Not Found", http.StatusNotFound)
		proxy.metrics.IncStatusCodes(404)
		return
	}

	// 转换请求头格式
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 选择上游服务
	upstream := proxy.router.GetUpstream(rule, headers)
	if upstream == nil || len(upstream.Addresses) == 0 {
		proxy.log.Errorf("没有可用的上游服务: %s", rule.Domain)
		http.Error(w, "503 Service Unavailable", http.StatusServiceUnavailable)
		proxy.metrics.IncStatusCodes(503)
		return
	}

	// 选择后端地址（简单的轮询）
	backendAddr := upstream.Addresses[0] // 简化处理，实际应该使用负载均衡算法

	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s:%d%s", backendAddr, upstream.Port, r.URL.Path)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// 创建转发请求
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		proxy.log.Errorf("创建转发请求失败: %v", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		proxy.metrics.IncStatusCodes(500)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 转发请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		proxy.log.Errorf("转发请求失败: %v", err)
		http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
		proxy.metrics.IncStatusCodes(502)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	if _, err := io.Copy(w, resp.Body); err != nil {
		proxy.log.Errorf("复制响应体失败: %v", err)
	}

	// 记录指标
	duration := time.Since(start)
	proxy.metrics.RecordLatency(duration)
	proxy.metrics.IncStatusCodes(resp.StatusCode)

	proxy.log.Debugf("HTTPS请求处理完成: %s -> %s, 耗时: %v", domain, targetURL, duration)
}

// findRouteForDomain 根据域名和路径查找路由规则
func (proxy *Proxy) findRouteForDomain(domain, path string) *RouteRule {
	return proxy.router.FindRouteByDomain(domain, path)
}

// AddCertificate 添加HTTPS证书
func (proxy *Proxy) AddCertificate(domain, certFile, keyFile string) error {
	return proxy.certManager.AddCertificate(domain, certFile, keyFile)
}

// RemoveCertificate 移除HTTPS证书
func (proxy *Proxy) RemoveCertificate(domain string) {
	proxy.certManager.RemoveCertificate(domain)
}

// Stop 停止代理服务器
func (proxy *Proxy) Stop() {
	proxy.cancel()
	proxy.client.CloseIdleConnections()

	// 优雅关闭HTTPS服务器
	if proxy.httpsServer != nil {
		proxy.httpsServer.Shutdown(context.Background())
	}

	proxy.log.Info("数据面代理服务器已停止")
}

// handleRequest 处理HTTP请求
func (proxy *Proxy) handleRequest(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	atomic.AddInt64(&proxy.connCount, 1)
	defer atomic.AddInt64(&proxy.connCount, -1)

	// 记录请求开始
	proxy.metrics.IncRequests()

	// 查找路由规则
	rule := proxy.router.FindRoute(ctx)
	if rule == nil {
		proxy.log.Warnf("未找到匹配的路由规则: %s%s", ctx.Host(), ctx.Path())
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString("404 Not Found")
		proxy.metrics.IncStatusCodes(404)
		return
	}

	// 提取请求头
	headers := make(map[string]string)
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	// 选择上游服务
	upstream := proxy.router.GetUpstream(rule, headers)
	if upstream == nil || len(upstream.Addresses) == 0 {
		proxy.log.Errorf("没有可用的上游服务: %s", rule.Domain)
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetBodyString("503 Service Unavailable")
		proxy.metrics.IncStatusCodes(503)
		return
	}

	// 选择后端地址（简单的轮询）
	backendAddr := upstream.Addresses[0] // 简化处理，实际应该使用负载均衡算法

	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s:%d%s", backendAddr, upstream.Port, ctx.Path())

	// 创建转发请求
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// 复制请求
	ctx.Request.CopyTo(req)
	req.SetRequestURI(targetURL)

	// 转发请求
	err := proxy.client.Do(req, resp)
	if err != nil {
		proxy.log.Errorf("转发请求失败: %v", err)
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetBodyString("502 Bad Gateway")
		proxy.metrics.IncStatusCodes(502)
		return
	}

	// 复制响应
	resp.CopyTo(&ctx.Response)

	// 记录指标
	duration := time.Since(start)
	proxy.metrics.RecordLatency(duration)
	proxy.metrics.IncStatusCodes(ctx.Response.StatusCode())

	proxy.log.Debugf("请求处理完成: %s -> %s, 耗时: %v", ctx.Host(), targetURL, duration)
}

// GetMetrics 获取监控指标
func (proxy *Proxy) GetMetrics() *Metrics {
	return proxy.metrics
}

// GetConnectionCount 获取当前连接数
func (proxy *Proxy) GetConnectionCount() int64 {
	return atomic.LoadInt64(&proxy.connCount)
}
