package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fredericobenevides/pge-rate-limiter/configs"
	"github.com/fredericobenevides/pge-rate-limiter/internal/handler"
	"github.com/fredericobenevides/pge-rate-limiter/internal/storage"
	"github.com/fredericobenevides/pge-rate-limiter/pkg/middleware"
	middlewares "github.com/fredericobenevides/pge-rate-limiter/pkg/middleware"
	"github.com/go-redis/redis/v8"
)

func main() {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	redisClient, rateLimitMiddleware := initDependencies(cfg)
	defer redisClient.Close()

	http.Handle("/", rateLimitMiddleware.Middleware(http.HandlerFunc(handler.HelloWorldHandler)))

	var httpPort = fmt.Sprintf(":%s", cfg.Port)

	log.Printf("Starting server on port: %s", cfg.Port)
	err = http.ListenAndServe(httpPort, nil)
	if err != nil {
		panic(err)
	}
}

func initDependencies(cfg *configs.RateLimiterConfig) (*redis.Client, *middleware.RateLimitMiddleware) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisHost,
	})

	var storageStrategy storage.StorageStrategy
	if cfg.RedisStorageEnabled {
		storageStrategy = storage.NewRedisStorageStrategy(client)
	} else {
		storageStrategy = storage.NewMemoryStorageStrategy()
	}

	rateLimitMiddleware := middlewares.NewRateLimitMiddleware(storageStrategy, cfg.RequestIPLimit, cfg.RequestTokenLimit, cfg.WindowDuration, cfg.BlockIPDuration, cfg.BlockTokenDuration)

	return client, rateLimitMiddleware
}
