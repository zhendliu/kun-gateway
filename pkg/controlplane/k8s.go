package controlplane

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sDiscovery K8s服务发现
type K8sDiscovery struct {
	client    *kubernetes.Clientset
	log       *logrus.Logger
	services  map[string]*ServiceInfo
	endpoints map[string]*EndpointInfo
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name      string
	Namespace string
	ClusterIP string
	Ports     []ServicePort
	Labels    map[string]string
	CreatedAt time.Time
}

// ServicePort 服务端口
type ServicePort struct {
	Name       string
	Port       int32
	TargetPort int32
	Protocol   string
}

// EndpointInfo 端点信息
type EndpointInfo struct {
	ServiceName string
	Namespace   string
	Addresses   []string // Pod IP列表
	Ports       []int32
	Ready       bool
	UpdatedAt   time.Time
}

// NewK8sDiscovery 创建K8s服务发现
func NewK8sDiscovery(log *logrus.Logger) (*K8sDiscovery, error) {
	// 尝试从集群内部获取配置
	config, err := rest.InClusterConfig()
	if err != nil {
		// 如果不在集群内，尝试从kubeconfig文件获取
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("无法获取K8s配置: %v", err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建K8s客户端失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &K8sDiscovery{
		client:    client,
		log:       log,
		services:  make(map[string]*ServiceInfo),
		endpoints: make(map[string]*EndpointInfo),
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start 启动服务发现
func (k *K8sDiscovery) Start() error {
	k.log.Info("启动K8s服务发现...")

	// 启动Service监听
	go k.watchServices()

	// 启动Endpoint监听
	go k.watchEndpoints()

	return nil
}

// Stop 停止服务发现
func (k *K8sDiscovery) Stop() {
	k.cancel()
	k.log.Info("K8s服务发现已停止")
}

// watchServices 监听Service变化
func (k *K8sDiscovery) watchServices() {
	for {
		select {
		case <-k.ctx.Done():
			return
		default:
			k.watchServicesOnce()
		}
	}
}

// watchServicesOnce 监听Service变化（单次）
func (k *K8sDiscovery) watchServicesOnce() {
	watcher, err := k.client.CoreV1().Services("").Watch(k.ctx, metav1.ListOptions{})
	if err != nil {
		k.log.Errorf("监听Service失败: %v", err)
		time.Sleep(5 * time.Second)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				k.log.Warn("Service监听通道已关闭，重新连接...")
				return
			}

			service, ok := event.Object.(*corev1.Service)
			if !ok {
				continue
			}

			key := fmt.Sprintf("%s/%s", service.Namespace, service.Name)

			switch event.Type {
			case watch.Added, watch.Modified:
				k.updateService(service)
				k.log.Infof("Service更新: %s, ClusterIP: %s", key, service.Spec.ClusterIP)
			case watch.Deleted:
				k.deleteService(key)
				k.log.Infof("Service删除: %s", key)
			}
		}
	}
}

// watchEndpoints 监听Endpoint变化
func (k *K8sDiscovery) watchEndpoints() {
	for {
		select {
		case <-k.ctx.Done():
			return
		default:
			k.watchEndpointsOnce()
		}
	}
}

// watchEndpointsOnce 监听Endpoint变化（单次）
func (k *K8sDiscovery) watchEndpointsOnce() {
	watcher, err := k.client.CoreV1().Endpoints("").Watch(k.ctx, metav1.ListOptions{})
	if err != nil {
		k.log.Errorf("监听Endpoint失败: %v", err)
		time.Sleep(5 * time.Second)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				k.log.Warn("Endpoint监听通道已关闭，重新连接...")
				return
			}

			endpoint, ok := event.Object.(*corev1.Endpoints)
			if !ok {
				continue
			}

			key := fmt.Sprintf("%s/%s", endpoint.Namespace, endpoint.Name)

			switch event.Type {
			case watch.Added, watch.Modified:
				k.updateEndpoint(endpoint)
				k.log.Infof("Endpoint更新: %s, 地址数量: %d", key, len(endpoint.Subsets))
			case watch.Deleted:
				k.deleteEndpoint(key)
				k.log.Infof("Endpoint删除: %s", key)
			}
		}
	}
}

// updateService 更新服务信息
func (k *K8sDiscovery) updateService(service *corev1.Service) {
	k.mu.Lock()
	defer k.mu.Unlock()

	key := fmt.Sprintf("%s/%s", service.Namespace, service.Name)

	ports := make([]ServicePort, 0, len(service.Spec.Ports))
	for _, port := range service.Spec.Ports {
		ports = append(ports, ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: port.TargetPort.IntVal,
			Protocol:   string(port.Protocol),
		})
	}

	k.services[key] = &ServiceInfo{
		Name:      service.Name,
		Namespace: service.Namespace,
		ClusterIP: service.Spec.ClusterIP,
		Ports:     ports,
		Labels:    service.Labels,
		CreatedAt: time.Now(),
	}
}

// deleteService 删除服务信息
func (k *K8sDiscovery) deleteService(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.services, key)
}

// updateEndpoint 更新端点信息
func (k *K8sDiscovery) updateEndpoint(endpoint *corev1.Endpoints) {
	k.mu.Lock()
	defer k.mu.Unlock()

	key := fmt.Sprintf("%s/%s", endpoint.Namespace, endpoint.Name)

	addresses := make([]string, 0)
	ports := make([]int32, 0)

	for _, subset := range endpoint.Subsets {
		// 只选择Ready状态的地址
		for _, address := range subset.Addresses {
			addresses = append(addresses, address.IP)
		}

		for _, port := range subset.Ports {
			ports = append(ports, port.Port)
		}
	}

	k.endpoints[key] = &EndpointInfo{
		ServiceName: endpoint.Name,
		Namespace:   endpoint.Namespace,
		Addresses:   addresses,
		Ports:       ports,
		Ready:       len(addresses) > 0,
		UpdatedAt:   time.Now(),
	}
}

// deleteEndpoint 删除端点信息
func (k *K8sDiscovery) deleteEndpoint(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.endpoints, key)
}

// GetServices 获取所有服务
func (k *K8sDiscovery) GetServices() map[string]*ServiceInfo {
	k.mu.RLock()
	defer k.mu.RUnlock()

	result := make(map[string]*ServiceInfo)
	for k, v := range k.services {
		result[k] = v
	}
	return result
}

// GetEndpoints 获取所有端点
func (k *K8sDiscovery) GetEndpoints() map[string]*EndpointInfo {
	k.mu.RLock()
	defer k.mu.RUnlock()

	result := make(map[string]*EndpointInfo)
	for k, v := range k.endpoints {
		result[k] = v
	}
	return result
}

// GetServiceEndpoints 获取指定服务的端点
func (k *K8sDiscovery) GetServiceEndpoints(namespace, serviceName string) *EndpointInfo {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key := fmt.Sprintf("%s/%s", namespace, serviceName)
	return k.endpoints[key]
}
