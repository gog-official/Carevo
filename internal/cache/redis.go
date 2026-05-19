package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
}

func New(addr string) *Store {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Store{client: client}
}

func (s *Store) Get(ctx context.Context, key string, dest any) (bool, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *Store) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *Store) Close() error {
	return s.client.Close()
}
