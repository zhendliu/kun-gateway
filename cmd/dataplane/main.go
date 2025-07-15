package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"kun-gateway/pkg/dataplane"

	"github.com/sirupsen/logrus"
)

var (
	port     = flag.Int("port", 80, "代理服务器监听端口")
	apiPort  = flag.Int("api-port", 8080, "API服务器监听端口")
	logLevel = flag.String("log-level", "info", "日志级别")
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

	log.Info("启动K8s流量网关数据面...")

	// 创建路由引擎
	router := dataplane.NewRouter(log)

	// 创建代理服务器
	proxy := dataplane.NewProxy(router, log)

	// 创建API服务器
	apiServer := dataplane.NewAPIServer(router, proxy, log)

	// 启动API服务器
	go func() {
		apiAddr := fmt.Sprintf(":%d", *apiPort)
		if err := apiServer.Start(apiAddr); err != nil {
			log.Fatalf("启动API服务器失败: %v", err)
		}
	}()

	// 启动代理服务器
	go func() {
		proxyAddr := fmt.Sprintf(":%d", *port)
		if err := proxy.Start(proxyAddr); err != nil {
			log.Fatalf("启动代理服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("收到中断信号，正在关闭服务...")

	// 优雅关闭
	proxy.Stop()

	log.Info("数据面服务已关闭")
}
