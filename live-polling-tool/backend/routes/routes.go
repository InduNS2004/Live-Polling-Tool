package routes

import (
	"github.com/gin-gonic/gin"
	"live-polling-tool/backend/config"
	"live-polling-tool/backend/controllers"
	"live-polling-tool/backend/middleware"
	"live-polling-tool/backend/realtime"
	"live-polling-tool/backend/repositories"
	"live-polling-tool/backend/services"
	"net/http"
)

func Setup(r *gin.Engine, cfg config.Config, repo *repositories.Repo, redis *realtime.Redis) {
	auth := &controllers.AuthController{S: &services.AuthService{Repo: repo, Secret: cfg.JWTSecret}, Secure: cfg.CookieSecure}
	poll := &controllers.PollController{Repo: repo, Redis: redis, Secure: cfg.CookieSecure}
	hub := &realtime.Hub{Redis: redis, Repo: repo}
	r.GET("/health", func(c *gin.Context) {
		dbOK := repo.DB.Client().Ping(c.Request.Context(), nil) == nil
		redisOK := redis != nil && redis.Ping(c.Request.Context()) == nil
		status := http.StatusOK
		if !dbOK {
			status = http.StatusServiceUnavailable
		}
		overall := "ok"
		if !dbOK || !redisOK {
			overall = "degraded"
		}
		c.JSON(status, gin.H{"status": overall, "database": map[bool]string{true: "connected", false: "disconnected"}[dbOK], "redis": map[bool]string{true: "connected", false: "disconnected"}[redisOK]})
	})
	a := r.Group("/api/auth")
	a.POST("/register", auth.Register)
	a.POST("/login", auth.Login)
	a.GET("/me", middleware.Auth(cfg.JWTSecret), auth.Me)
	a.POST("/logout", func(c *gin.Context) {
		c.SetCookie("auth_token", "", -1, "/", "", cfg.CookieSecure, true)
		c.Status(204)
	})
	pub := r.Group("/api/polls")
	pub.GET("/:id", poll.Get)
	pub.GET("/:id/results", poll.Results)
	pub.POST("/:id/vote", poll.Vote)
	pub.GET("/:id/live", hub.Serve)
	protected := r.Group("/api/polls", middleware.Auth(cfg.JWTSecret))
	protected.GET("", poll.List)
	protected.POST("", poll.Create)
	protected.PATCH("/:id/close", poll.Close)
	protected.DELETE("/:id", poll.Delete)
}
