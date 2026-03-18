package main

import (
	"go-api/internal/cache"
	"go-api/internal/config/db"
	apilogger "go-api/internal/logger"
	middleware_logger "go-api/internal/middleware/logger"
	"go-api/internal/nats"
	"go-api/internal/routes"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-api/docs"
)

// @title Go API
// @version 1.0
// @description API for user management, payments and resumes
// @host localhost:8080
// @BasePath /

func main() {
	go func() {
		log.Info().Msg("pprof on :6060")
		http.ListenAndServe("localhost:6060", nil)
	}()

	apilogger.Init()
	gin.SetMode(gin.ReleaseMode)

	if err := cache.Init(); err != nil {
		log.Warn().Err(err).Msg("redis not available, caching disabled")
	} else {
		log.Info().Msg("redis connected")
	}

	if err := nats.Init(); err != nil {
		log.Warn().Err(err).Msg("nats not available, queue disabled")
	} else {
		log.Info().Msg("nats connected")
		defer nats.Close()
	}

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware_logger.Logger())

	r.SetTrustedProxies([]string{"127.0.0.1"})
	db.Init()

	routes.SetupRoutes(r)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Info().Str("addr", ":8080").Msg("server started")

	if err := r.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
