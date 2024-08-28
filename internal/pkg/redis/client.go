package redis

import (
	"context"
	"errors"
	"strings"

	"github.com/go-redis/redis/v8"
)

func New(ctx context.Context, redisConfig Config) (redis.UniversalClient, error) {
	redisAddresses := strings.Split(redisConfig.Addresses, ",")
	if len(redisAddresses) == 0 {
		return nil, errors.New("redis host is empty")
	}

	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Password: redisConfig.Password,
		Addrs:    redisAddresses,
		DB:       redisConfig.DB,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return redisClient, nil
}
