package dataplane

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// APIServer 数据面API服务器
type APIServer struct {
	router *Router
	proxy  *Proxy
	log    *logrus.Logger
}

// NewAPIServer 创建API服务器
func NewAPIServer(router *Router, proxy *Proxy, log *logrus.Logger) *APIServer {
	return &APIServer{
		router: router,
		proxy:  proxy,
		log:    log,
	}
}

// Start 启动API服务器
func (api *APIServer) Start(addr string) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 路由管理API
	r.PUT("/api/v1/routes", api.updateRoutes)
	r.GET("/api/v1/routes", api.getRoutes)

	// 监控指标API
	r.GET("/api/v1/metrics", api.getMetrics)
	r.GET("/api/v1/health", api.healthCheck)

	api.log.Infof("数据面API服务器启动，监听地址: %s", addr)
	return r.Run(addr)
}

// RouteUpdateRequest 路由更新请求
type RouteUpdateRequest struct {
	Routes []*RouteRule `json:"routes"`
}

// RouteUpdateResponse 路由更新响应
type RouteUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// updateRoutes 更新路由规则
func (api *APIServer) updateRoutes(c *gin.Context) {
	var req RouteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	api.log.Infof("收到路由更新请求，规则数量: %d", len(req.Routes))

	// 原子更新路由规则
	api.router.UpdateRules(req.Routes)

	c.JSON(http.StatusOK, RouteUpdateResponse{
		Success: true,
		Message: "路由规则更新成功",
		Count:   len(req.Routes),
	})
}

// getRoutes 获取当前路由规则
func (api *APIServer) getRoutes(c *gin.Context) {
	table := api.router.rules.Load().(*RouteTable)

	rules := make([]*RouteRule, 0, len(table.Rules))
	for _, rule := range table.Rules {
		rules = append(rules, rule)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"routes":  rules,
		"count":   len(rules),
	})
}

// getMetrics 获取监控指标
func (api *APIServer) getMetrics(c *gin.Context) {
	metrics := api.proxy.GetMetrics()
	stats := metrics.GetStats()

	// 添加连接数
	stats["connection_count"] = api.proxy.GetConnectionCount()
	stats["timestamp"] = time.Now().Unix()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": stats,
	})
}

// healthCheck 健康检查
func (api *APIServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "dataplane",
	})
}
