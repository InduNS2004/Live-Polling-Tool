package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"live-polling-tool/backend/config"
	"live-polling-tool/backend/database"
	"live-polling-tool/backend/realtime"
	"live-polling-tool/backend/repositories"
	"live-polling-tool/backend/routes"
	"log"
	"time"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	m, err := database.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatal("mongodb: ", err)
	}
	defer m.Close(ctx)
	repo := repositories.New(m.DB)
	if err = repo.Init(ctx); err != nil {
		log.Fatal("indexes: ", err)
	}
	var redis *realtime.Redis
	if cfg.RedisURL != "" {
		redis, err = realtime.NewRedis(cfg.RedisURL)
		if err != nil {
			log.Printf("redis config: %v", err)
		} else if e := redis.Ping(ctx); e != nil {
			log.Printf("redis unavailable: %v", e)
			redis = nil
		} else {
			log.Println("redis connected")
		}
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{AllowOrigins: []string{cfg.FrontendURL}, AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Authorization"}, AllowCredentials: true, MaxAge: 12 * time.Hour}))
	routes.Setup(r, cfg, repo, redis)
	log.Printf("live polling backend listening on :%s", cfg.Port)
	if err = r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
