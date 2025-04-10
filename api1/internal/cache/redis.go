package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	customerrors "github.com/nikita89756/testEffectiveMobile/internal/errors"
	"github.com/nikita89756/testEffectiveMobile/internal/model"
	"github.com/nikita89756/testEffectiveMobile/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const(
	defaultTTL = 5 * time.Hour
)

type Cache interface {
	SetPersonWithTTL(ctx context.Context, name string, person model.PersonStats) error
	GetPerson(ctx context.Context, name string) (*model.PersonStats, error)
}


type RedisClient struct {
	client *redis.Client
	logger logger.Logger
}

func NewRedisClient(addr, password string, db int,logger logger.Logger) (Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}

	return &RedisClient{client: rdb,logger : logger,}, nil
}
func (r *RedisClient) SetPersonWithTTL(ctx context.Context, name string, person model.PersonStats, ) error {
	key := name
	value, err := json.Marshal(person)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации объекта Person в JSON: %w", err)
	}
	err = r.client.SetEx(ctx, key, value, defaultTTL).Err()
	if err != nil {
		return fmt.Errorf("ошибка при записи ключа %s с TTL: %w", key, err)
	}
	r.logger.Info("Ключ записан в Redis", zap.String("key", key) )
	return nil
}

// GetPerson - получить PersonStats по имени из Redis, если ключ не найден, то возвращается ErrKeyNotFound
func (r *RedisClient) GetPerson(ctx context.Context, name string) (*model.PersonStats, error) {
	key := name
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, customerrors.ErrKeyNotFound 
	} else if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ключа %s: %w", key, err)
	}

	var person model.PersonStats
	err = json.Unmarshal([]byte(value), &person)
	if err != nil {
		return nil, fmt.Errorf("ошибка при десериализации JSON в объект Person: %w", err)
	}
	return &person, nil
}
