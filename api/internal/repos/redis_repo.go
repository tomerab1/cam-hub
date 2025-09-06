package repos

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIface interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

type RedisRepo struct {
	Rdb    *redis.Client
	Logger *slog.Logger
}

func (repo *RedisRepo) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	status := repo.Rdb.Set(ctx, key, value, ttl)
	repo.Logger.Info("Cache set", "key", key, "value", value)
	return status.Err()
}

func (repo *RedisRepo) Get(ctx context.Context, key string) (string, error) {
	cmd := repo.Rdb.Get(ctx, key)
	repo.Logger.Info("Cache get", "key", key)

	val, err := cmd.Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (repo *RedisRepo) Del(ctx context.Context, keys ...string) error {
	status := repo.Rdb.Del(ctx, keys...)
	repo.Logger.Info("Cache del", "keys", keys)
	return status.Err()
}
