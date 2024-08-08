package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	syncBlockMetadataKey = "syncBlockMetadata"
)

type SyncBlockMetadataRepository struct {
	prefix      string
	redisClient redis.UniversalClient
}

func NewSyncBlockMetadataRepository(prefix string, redisClient redis.UniversalClient) *SyncBlockMetadataRepository {
	return &SyncBlockMetadataRepository{
		redisClient: redisClient,
		prefix:      prefix,
	}
}
func (r *SyncBlockMetadataRepository) GetHead(ctx context.Context) (string, error) {
	result, err := r.redisClient.Get(ctx, r.getKey()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}

	return result, nil
}

func (r *SyncBlockMetadataRepository) SetHead(ctx context.Context, blockHash string) error {
	err := r.redisClient.Set(ctx, r.getKey(), blockHash, -1).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *SyncBlockMetadataRepository) getKey() string {
	return fmt.Sprintf("%s:%s", r.prefix, syncBlockMetadataKey)
}
