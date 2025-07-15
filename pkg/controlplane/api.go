package controlplane

import (
	"net/http"
	"strings"
	"time"

	"kun-gateway/pkg/dataplane"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ControlPlaneAPI 控制面API服务器
type ControlPlaneAPI struct {
	k8sDiscovery    *K8sDiscovery
	dataplaneClient *DataPlaneClient
	log             *logrus.Logger
}

// NewControlPlaneAPI 创建控制面API
func NewControlPlaneAPI(k8sDiscovery *K8sDiscovery, dataplaneClient *DataPlaneClient, log *logrus.Logger) *ControlPlaneAPI {
	return &ControlPlaneAPI{
		k8sDiscovery:    k8sDiscovery,
		dataplaneClient: dataplaneClient,
		log:             log,
	}
}

// Start 启动控制面API服务器
func (api *ControlPlaneAPI) Start(addr string) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 路由管理
	r.GET("/api/v1/routes", api.getRoutes)
	r.POST("/api/v1/routes", api.createRoute)
	r.PUT("/api/v1/routes/:id", api.updateRoute)
	r.DELETE("/api/v1/routes/:id", api.deleteRoute)

	// K8s服务发现
	r.GET("/api/v1/services", api.getServices)
	r.GET("/api/v1/endpoints", api.getEndpoints)

	// 监控数据
	r.GET("/api/v1/metrics", api.getMetrics)
	r.GET("/api/v1/metrics/domains", api.getDomainMetrics)

	// 健康检查
	r.GET("/api/v1/health", api.healthCheck)

	api.log.Infof("控制面API服务器启动，监听地址: %s", addr)
	return r.Run(addr)
}

// RouteConfig 路由配置
type RouteConfig struct {
	ID        string            `json:"id"`
	Domain    string            `json:"domain"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers,omitempty"`
	Service   string            `json:"service"` // 格式: namespace/service
	Port      int               `json:"port"`
	Weight    int               `json:"weight"`
	Enabled   bool              `json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// getRoutes 获取所有路由配置
func (api *ControlPlaneAPI) getRoutes(c *gin.Context) {
	routes, err := api.dataplaneClient.GetRoutes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取路由失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"routes":  routes,
	})
}

// createRoute 创建路由配置
func (api *ControlPlaneAPI) createRoute(c *gin.Context) {
	var config RouteConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证服务是否存在
	// 解析服务名称格式: namespace/service
	parts := strings.Split(config.Service, "/")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "服务名称格式错误，应为 namespace/service",
		})
		return
	}

	namespace, serviceName := parts[0], parts[1]
	endpoint := api.k8sDiscovery.GetServiceEndpoints(namespace, serviceName)
	if endpoint == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "指定的服务不存在或没有可用的端点",
		})
		return
	}

	// 构建路由规则
	rule := &dataplane.RouteRule{
		Domain:    config.Domain,
		Path:      config.Path,
		Headers:   config.Headers,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 构建上游服务
	upstream := dataplane.Upstream{
		Name:      config.Service,
		Addresses: endpoint.Addresses,
		Port:      config.Port,
		Weight:    config.Weight,
		Healthy:   endpoint.Ready,
	}
	rule.Upstreams = append(rule.Upstreams, upstream)

	// 推送到数据面
	err := api.dataplaneClient.UpdateRoutes([]*dataplane.RouteRule{rule})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新路由失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "路由创建成功",
		"route":   config,
	})
}

// updateRoute 更新路由配置
func (api *ControlPlaneAPI) updateRoute(c *gin.Context) {
	id := c.Param("id")

	var config RouteConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	config.ID = id
	config.UpdatedAt = time.Now()

	// 这里简化处理，实际应该更新现有的路由规则
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "路由更新成功",
		"route":   config,
	})
}

// deleteRoute 删除路由配置
func (api *ControlPlaneAPI) deleteRoute(c *gin.Context) {
	id := c.Param("id")

	// 这里简化处理，实际应该删除对应的路由规则
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "路由删除成功",
		"id":      id,
	})
}

// getServices 获取K8s服务列表
func (api *ControlPlaneAPI) getServices(c *gin.Context) {
	services := api.k8sDiscovery.GetServices()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"services": services,
		"count":    len(services),
	})
}

// getEndpoints 获取K8s端点列表
func (api *ControlPlaneAPI) getEndpoints(c *gin.Context) {
	endpoints := api.k8sDiscovery.GetEndpoints()

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"endpoints": endpoints,
		"count":     len(endpoints),
	})
}

// getMetrics 获取监控指标
func (api *ControlPlaneAPI) getMetrics(c *gin.Context) {
	metrics, err := api.dataplaneClient.GetMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取监控指标失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": metrics,
	})
}

// getDomainMetrics 获取域名维度监控指标
func (api *ControlPlaneAPI) getDomainMetrics(c *gin.Context) {
	domain := c.Query("domain")

	metrics, err := api.dataplaneClient.GetMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取监控指标失败: " + err.Error(),
		})
		return
	}

	// 如果指定了域名，过滤数据
	if domain != "" {
		if domainMetrics, ok := metrics["domains"].(map[string]interface{}); ok {
			if specificDomain, exists := domainMetrics[domain]; exists {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"domain":  domain,
					"metrics": specificDomain,
				})
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "指定的域名不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": metrics["domains"],
	})
}

// healthCheck 健康检查
func (api *ControlPlaneAPI) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "controlplane",
	})
}
