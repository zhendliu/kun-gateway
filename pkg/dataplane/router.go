package dataplane

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

// RouteRule 路由规则
type RouteRule struct {
	Domain    string            `json:"domain"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Upstreams []Upstream        `json:"upstreams"`
	Weight    map[string]int    `json:"weight"` // 流量权重分配
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Upstream 上游服务
type Upstream struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"` // Pod IP列表
	Port      int      `json:"port"`
	Weight    int      `json:"weight"`
	Healthy   bool     `json:"healthy"`
}

// Router 路由引擎
type Router struct {
	rules atomic.Value // *RouteTable
	log   *logrus.Logger
}

// RouteTable 路由表
type RouteTable struct {
	Rules map[string]*RouteRule // key: domain+path
}

// NewRouter 创建路由引擎
func NewRouter(log *logrus.Logger) *Router {
	r := &Router{log: log}
	r.rules.Store(&RouteTable{Rules: make(map[string]*RouteRule)})
	return r
}

// UpdateRules 原子更新路由规则
func (r *Router) UpdateRules(rules []*RouteRule) {
	newTable := &RouteTable{Rules: make(map[string]*RouteRule)}

	for _, rule := range rules {
		key := fmt.Sprintf("%s%s", rule.Domain, rule.Path)
		newTable.Rules[key] = rule
	}

	r.rules.Store(newTable)
	r.log.Infof("路由规则已更新，共 %d 条规则", len(rules))
}

// FindRoute 查找匹配的路由规则
func (r *Router) FindRoute(ctx *fasthttp.RequestCtx) *RouteRule {
	table := r.rules.Load().(*RouteTable)
	host := string(ctx.Host())
	path := string(ctx.Path())

	// 精确匹配
	key := host + path
	if rule, exists := table.Rules[key]; exists {
		return rule
	}

	// 域名匹配
	for _, rule := range table.Rules {
		if rule.Domain == host {
			return rule
		}
	}

	return nil
}

// GetUpstream 根据权重选择上游服务
func (r *Router) GetUpstream(rule *RouteRule, headers map[string]string) *Upstream {
	// 检查Header路由
	if rule.Headers != nil {
		for headerKey, headerValue := range rule.Headers {
			if clientValue := headers[headerKey]; clientValue == headerValue {
				// 根据Header值选择特定的Upstream
				for _, upstream := range rule.Upstreams {
					if upstream.Name == headerValue {
						return &upstream
					}
				}
			}
		}
	}

	// 权重分配
	if len(rule.Upstreams) == 0 {
		return nil
	}

	// 简单的轮询选择（实际项目中可以使用更复杂的负载均衡算法）
	totalWeight := 0
	for _, upstream := range rule.Upstreams {
		totalWeight += upstream.Weight
	}

	if totalWeight == 0 {
		return &rule.Upstreams[0]
	}

	// 这里简化处理，实际应该使用更精确的权重算法
	selected := time.Now().UnixNano() % int64(totalWeight)
	currentWeight := 0

	for _, upstream := range rule.Upstreams {
		currentWeight += upstream.Weight
		if int64(currentWeight) > selected {
			return &upstream
		}
	}

	return &rule.Upstreams[0]
}
