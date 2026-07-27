package installer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	shutdownCh chan struct{}
	wizardSrv  *http.Server
)

// Shutdown 安全关闭安装向导 HTTP 服务
func Shutdown() {
	if shutdownCh != nil {
		select {
		case <-shutdownCh:
			// already closed
		default:
			close(shutdownCh)
		}
	}
	if wizardSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		wizardSrv.Shutdown(ctx)
	}
}

// Run 启动安装向导 HTTP 服务（临时端口 9898）
// 安装完成后自动关闭，调用方可继续正常启动
func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	RegisterRoutes(r)

	shutdownCh = make(chan struct{})
	wizardSrv = &http.Server{
		Addr:         ":9898",
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 尝试获取本机 IP
	ip := ""
	if host, err := os.Hostname(); err == nil {
		ip = host
	}

	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("  Xt-PMS 安装向导")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("  👉 http://localhost:9898")
	if ip != "" {
		fmt.Printf("  👉 http://%s:9898\n", ip)
	}
	fmt.Println()
	fmt.Println("  安装完成后将自动启动服务")
	fmt.Println("==========================================")
	fmt.Println()

	// 在 goroutine 中启动，主 goroutine 等待关闭信号
	go func() {
		if err := wizardSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println()
			fmt.Println("==========================================")
			fmt.Printf("  安装向导启动失败：端口 9898 已被占用\n")
			fmt.Println("  请检查是否有其他 xt-pms 实例正在运行")
			fmt.Println("  停止现有服务后重新启动即可")
			fmt.Println("==========================================")
			fmt.Println()
			os.Exit(1)
		}
	}()

	// 阻塞直到安装完成调用 Shutdown()
	<-shutdownCh
	Shutdown()

	fmt.Println()
	fmt.Println("✓ 安装完成，正在启动服务...")
	fmt.Println()
}
