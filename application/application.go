package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sine-io/sinx/api/handler"
	"github.com/sine-io/sinx/api/router"
	userAppService "github.com/sine-io/sinx/application/user/service"
	userDomainService "github.com/sine-io/sinx/domain/user/service"
	"github.com/sine-io/sinx/infra/database"
	"github.com/sine-io/sinx/infra/migration"
	userRepo "github.com/sine-io/sinx/infra/repository"
	"github.com/sine-io/sinx/pkg/config"
	"github.com/sine-io/sinx/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Application struct {
	server *http.Server
	db     *gorm.DB
}

type Dependencies struct {
	DB *gorm.DB
}

func Init(ctx context.Context) (*Application, error) {
	// 初始化基础设施
	deps, err := initInfrastructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init infrastructure: %w", err)
	}

	// 初始化服务层
	services, err := initServices(deps)
	if err != nil {
		return nil, fmt.Errorf("failed to init services: %w", err)
	}

	// 初始化处理器
	handlers := initHandlers(services)

	// 初始化HTTP服务器
	server := initHTTPServer(handlers)

	return &Application{
		server: server,
		db:     deps.DB,
	}, nil
}

func initInfrastructure(ctx context.Context) (*Dependencies, error) {
	// 初始化数据库
	db, err := database.NewPostgresDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 执行数据库迁移
	if err := migration.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Dependencies{
		DB: db,
	}, nil
}

type Services struct {
	UserAppService *userAppService.UserApplicationService
	// 预留: Role/Menu/RBAC 服务
}

func initServices(deps *Dependencies) (*Services, error) {
	// 初始化仓储层
	userRepository := userRepo.NewUserRepository(deps.DB)

	// TODO: 初始化角色 / 菜单 / RBAC 仓储与服务

	// 初始化领域服务层
	userDomainSvc := userDomainService.NewUserDomainService(userRepository)

	// 初始化应用服务层
	userAppSvc := userAppService.NewUserApplicationService(userDomainSvc)

	return &Services{
		UserAppService: userAppSvc,
	}, nil
}

type Handlers struct {
	UserHandler *handler.UserHandler
	// 预留: Role/Menu/RBAC 处理器
}

func initHandlers(services *Services) *Handlers {
	return &Handlers{
		UserHandler: handler.NewUserHandler(services.UserAppService),
	}
}

func initHTTPServer(handlers *Handlers) *http.Server {
	cfg := config.Get()

	// 设置Gin模式
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 设置路由
	router.SetupRoutes(r, handlers.UserHandler)

	return &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: r,
	}
}

func (app *Application) StartHTTPServer() error {
	cfg := config.Get()
	logger.Info("Starting HTTP server", "addr", cfg.ListenAddr)
	err := app.server.ListenAndServe()
	// http.ErrServerClosed 是正常的关闭流程，不作为错误返回
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (app *Application) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down HTTP server...")

	// 关闭HTTP服务器
	if err := app.server.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown HTTP server", "error", err)
		return err
	}

	// 关闭数据库连接
	if app.db != nil {
		sqlDB, err := app.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return nil
}
