package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
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
		DB:       cfg.Index,
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

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {

	key = fmt.Sprintf("%s:%s", c.prefix, key)
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrKeyNotFound
		}
		return err
	}

	err = msgpack.Unmarshal([]byte(value), &dest)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	key = fmt.Sprintf("%s:%s", c.prefix, key)

	data, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	err = c.client.Set(ctx, key, data, expiry).Err()
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
