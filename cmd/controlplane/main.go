package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"kun-gateway/pkg/controlplane"

	"github.com/sirupsen/logrus"
)

var (
	port         = flag.Int("port", 9090, "控制面API服务器监听端口")
	dataplaneURL = flag.String("dataplane-url", "http://localhost:8080", "数据面API地址")
	logLevel     = flag.String("log-level", "info", "日志级别")
)

func main() {
	flag.Parse()

	// 配置日志
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	log := logrus.New()
	log.SetLevel(level)

	log.Info("启动K8s流量网关控制面...")

	// 创建K8s服务发现
	k8sDiscovery, err := controlplane.NewK8sDiscovery(log)
	if err != nil {
		log.Fatalf("创建K8s服务发现失败: %v", err)
	}

	// 启动K8s服务发现
	if err := k8sDiscovery.Start(); err != nil {
		log.Fatalf("启动K8s服务发现失败: %v", err)
	}

	// 创建数据面客户端
	dataplaneClient := controlplane.NewDataPlaneClient(*dataplaneURL, log)

	// 检查数据面连接
	if err := dataplaneClient.HealthCheck(); err != nil {
		log.Warnf("数据面连接检查失败: %v", err)
	} else {
		log.Info("数据面连接正常")
	}

	// 创建控制面API服务器
	apiServer := controlplane.NewControlPlaneAPI(k8sDiscovery, dataplaneClient, log)

	// 启动控制面API服务器
	go func() {
		apiAddr := fmt.Sprintf(":%d", *port)
		if err := apiServer.Start(apiAddr); err != nil {
			log.Fatalf("启动控制面API服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到中断信号，正在关闭服务...")

	// 优雅关闭
	k8sDiscovery.Stop()

	log.Info("控制面服务已关闭")
}
