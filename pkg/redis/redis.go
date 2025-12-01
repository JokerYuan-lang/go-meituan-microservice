package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerYuan-lang/go-meituan-microservice/pkg/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var RedisClient *redis.Client
var ctx = context.Background()

func InitRedis() {
	cfg := config.Cfg.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		zap.L().Fatal("redis 连接失败", zap.Error(err))
	}
	RedisClient = client
	zap.L().Info("Redis 初始化成功")
}

func Set(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

func Del(key string) error {
	return RedisClient.Del(ctx, key).Err()
}
