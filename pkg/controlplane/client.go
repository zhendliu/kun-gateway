package controlplane

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kun-gateway/pkg/dataplane"

	"github.com/sirupsen/logrus"
)

// DataPlaneClient 数据面客户端
type DataPlaneClient struct {
	baseURL string
	client  *http.Client
	log     *logrus.Logger
}

// NewDataPlaneClient 创建数据面客户端
func NewDataPlaneClient(baseURL string, log *logrus.Logger) *DataPlaneClient {
	return &DataPlaneClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// UpdateRoutes 更新路由规则
func (c *DataPlaneClient) UpdateRoutes(rules []*dataplane.RouteRule) error {
	url := fmt.Sprintf("%s/api/v1/routes", c.baseURL)

	request := dataplane.RouteUpdateRequest{
		Routes: rules,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化路由规则失败: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("更新路由失败，状态码: %d", resp.StatusCode)
	}

	var response dataplane.RouteUpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("更新路由失败: %s", response.Message)
	}

	c.log.Infof("路由更新成功: %s", response.Message)
	return nil
}

// GetRoutes 获取路由规则
func (c *DataPlaneClient) GetRoutes() ([]*dataplane.RouteRule, error) {
	url := fmt.Sprintf("%s/api/v1/routes", c.baseURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取路由失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取路由失败，状态码: %d", resp.StatusCode)
	}

	var response struct {
		Success bool                   `json:"success"`
		Routes  []*dataplane.RouteRule `json:"routes"`
		Count   int                    `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("获取路由失败")
	}

	return response.Routes, nil
}

// GetMetrics 获取监控指标
func (c *DataPlaneClient) GetMetrics() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/metrics", c.baseURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取监控指标失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取监控指标失败，状态码: %d", resp.StatusCode)
	}

	var response struct {
		Success bool                   `json:"success"`
		Metrics map[string]interface{} `json:"metrics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("获取监控指标失败")
	}

	return response.Metrics, nil
}

// HealthCheck 健康检查
func (c *DataPlaneClient) HealthCheck() error {
	url := fmt.Sprintf("%s/api/v1/health", c.baseURL)

	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("健康检查失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("数据面服务不健康，状态码: %d", resp.StatusCode)
	}

	return nil
}
