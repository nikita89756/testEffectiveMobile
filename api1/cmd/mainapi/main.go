package main

import (
	"log"
	"time"

	services "github.com/nikita89756/testEffectiveMobile/internal/apis"
	cache "github.com/nikita89756/testEffectiveMobile/internal/cache"
	"github.com/nikita89756/testEffectiveMobile/internal/handlers"
	"github.com/nikita89756/testEffectiveMobile/internal/server"
	"github.com/nikita89756/testEffectiveMobile/internal/storage"
	config "github.com/nikita89756/testEffectiveMobile/pkg/config"
	logger "github.com/nikita89756/testEffectiveMobile/pkg/logger"
	"go.uber.org/zap"

	_ "github.com/nikita89756/testEffectiveMobile/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)
const(timeout = 5*time.Second)


// @title Effective Mobile API
// @version 1.0
// @host 0.0.0.0:8080
// @BasePath /api
func main() {
	cfg := config.InitConfig()

	logger,err := logger.InitLogger(true, "", cfg.LogLevel)

	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Инициализация логгера завершена")

	db:= storage.NewStorage(cfg.DatabaseConnection,logger, cfg.DBTimeout)

	err = db.Migrate(cfg.MigrationDir)
	if err != nil {
		logger.Error("ошибка миграции базы данных", zap.Error(err))
		return
	}
	cache,err := cache.NewRedisClient(cfg.Cache.Address, cfg.Cache.Password, cfg.Cache.Db,logger)
	if err != nil {
		logger.Error("ошибка создания клиента Redis")
		return
	}

	addon := services.NewAddonService(timeout,logger)

	handler := handlers.NewHandler(db, logger,addon,cache)

	// TODO server initializer

	server := server.New(cfg.Server.Host, cfg.Server.Port, handler)

	router:=server.CreateRoute()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	err=router.Run()

	if err != nil {
		log.Fatal(err)
	}
	// defer func() {
	// 	if err := server.Shutdown(); err != nil {
	// 		logger.Error("ошибка завершения работы сервера", err)
	// 	}
	// }()

}