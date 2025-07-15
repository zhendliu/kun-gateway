package dataplane

import (
	"context"
	"fmt"
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
		router:  router,
		log:     log,
		metrics: NewMetrics(),
		client:  client,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动代理服务器
func (proxy *Proxy) Start(addr string) error {
	proxy.log.Infof("启动数据面代理服务器，监听地址: %s", addr)

	// 使用fasthttp启动服务器
	return fasthttp.ListenAndServe(addr, proxy.handleRequest)
}

// Stop 停止代理服务器
func (proxy *Proxy) Stop() {
	proxy.cancel()
	proxy.client.CloseIdleConnections()
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
