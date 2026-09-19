package config

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	MongoURI     string
	MongoDB      string
	RedisURL     string
	JWTSecret    string
	FrontendURL  string
	CookieSecure bool
}

func Load() Config {
	return Config{
		Port:         env("PORT", "8080"),
		MongoURI:     env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:      env("MONGO_DB", "live_polling"),
		RedisURL:     env("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:    env("JWT_SECRET", "change-me-in-development"),
		FrontendURL:  env("FRONTEND_URL", "http://localhost:5173"),
		CookieSecure: strings.EqualFold(env("COOKIE_SECURE", "false"), "true"),
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
