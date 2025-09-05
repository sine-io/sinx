package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sine-io/sinx/pkg/config"
	"github.com/sine-io/sinx/pkg/logger"
)

var rdb *redis.Client

func InitRedis() error {
	cfg := config.Get()
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("redis_ping_failed", "err", err)
		// 可容忍: 返回 nil 表示继续运行(使用内存缓存 fallback) 也可直接返回错误
		return err
	}
	logger.Info("redis_connected")
	return nil
}

func GetRedis() *redis.Client { return rdb }
