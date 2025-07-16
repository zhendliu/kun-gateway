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

// GetCertificates 获取证书列表
func (client *DataPlaneClient) GetCertificates() ([]string, error) {
	resp, err := client.client.Get(client.baseURL + "/api/v1/certificates")
	if err != nil {
		return nil, fmt.Errorf("请求证书列表失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取证书列表失败，状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success      bool     `json:"success"`
		Certificates []string `json:"certificates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析证书列表响应失败: %v", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("获取证书列表失败")
	}

	return result.Certificates, nil
}

// AddCertificate 添加证书
func (client *DataPlaneClient) AddCertificate(domain, certFile, keyFile string) error {
	data := map[string]string{
		"domain":    domain,
		"cert_file": certFile,
		"key_file":  keyFile,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化证书数据失败: %v", err)
	}

	req, err := http.NewRequest("POST", client.baseURL+"/api/v1/certificates", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建添加证书请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.client.Do(req)
	if err != nil {
		return fmt.Errorf("添加证书请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("添加证书失败，状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析添加证书响应失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("添加证书失败: %s", result.Message)
	}

	return nil
}

// RemoveCertificate 移除证书
func (client *DataPlaneClient) RemoveCertificate(domain string) error {
	req, err := http.NewRequest("DELETE", client.baseURL+"/api/v1/certificates/"+domain, nil)
	if err != nil {
		return fmt.Errorf("创建删除请求失败: %v", err)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return fmt.Errorf("移除证书请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("移除证书失败，状态码: %d", resp.StatusCode)
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析移除证书响应失败: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("移除证书失败: %s", result.Message)
	}

	return nil
}
