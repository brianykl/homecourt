package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Something interface {
	Store() error
	Get() ([]string, error)
	Delete() error
}

type redisSomethingManager struct {
	client *redis.Client
}

// Delete implements Something.
func (r *redisSomethingManager) Delete() error {
	panic("unimplemented")
}

// Get implements Something.
func (r *redisSomethingManager) Get() ([]string, error) {
	panic("unimplemented")
}

// Store implements Something.
func (r *redisSomethingManager) Store() error {
	panic("unimplemented")
}

func NewSomethingManager(addr string) (Something, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &redisSomethingManager{client: client}, nil

}
