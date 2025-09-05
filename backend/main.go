package main

// @title        Sinx 用户认证系统 API
// @version      1.0
// @description  基于 Gin 的用户认证与授权服务接口文档。
// @contact.name Sineio
// @BasePath     /
// @schemes      http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description 使用格式: Bearer <token>

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sine-io/sinx/application"
	"github.com/sine-io/sinx/pkg/config"
	"github.com/sine-io/sinx/pkg/logger"

	// swagger docs（生成后自动导入，不生成也不会影响编译）
	_ "github.com/sine-io/sinx/docs"
)

func main() {
	// 设置崩溃输出
	setCrashOutput()

	// 加载环境变量
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}

	// 初始化日志
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 创建上下文
	ctx := context.Background()

	// 初始化应用
	app, err := application.Init(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize application", "error", err)
	}

	// 启动HTTP服务器
	go func() {
		if err := app.StartHTTPServer(); err != nil {
			// 仅在非正常关闭时才输出 Fatal
			logger.Fatal("HTTP server failed", "error", err)
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
	logger.Sync()
}

func setCrashOutput() {
	// 可以在这里设置崩溃日志输出文件
	// 当前简单处理，实际项目中可以输出到文件
}
