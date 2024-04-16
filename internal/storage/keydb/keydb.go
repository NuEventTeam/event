package keydb

import (
	"context"
	"errors"
	"fmt"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

var (
	ErrKeyNotFound = fmt.Errorf("key does not exists")
)

type Cache struct {
	client *redis.Client
	prefix string
}

func New(ctx context.Context, cfg config.Cache) *Cache {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		Password: cfg.Password},
	)

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	return &Cache{
		client: client,
		prefix: cfg.Prefix,
	}
}

func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {

	key = fmt.Sprintf("%s:%s", c.prefix, key)
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	return value, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	key = fmt.Sprintf("%s:%s", c.prefix, key)

	err := c.client.Set(ctx, key, value, expiry).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) Del(ctx context.Context, key string) error {
	key = fmt.Sprintf("%s:%s", c.prefix, key)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}
