package redis

import (
	"context"
	"errors"
	"time"
	"urlshort"

	"github.com/redis/go-redis/v9"
)

type client struct {
	*redis.Client
	expirationMinutes int
}

type Options struct {
	Addr              string
	Password          string
	DB                int
	ExpirationMinutes int
}

func New(opts *Options) *client {
	r := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})
	return &client{
		Client:            r,
		expirationMinutes: opts.ExpirationMinutes,
	}
}

func (c *client) Save(ctx context.Context, key string, url string) error {
	cmd := c.Client.SetNX(ctx, key, url, time.Duration(c.expirationMinutes)*time.Minute)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	cmd := c.Client.Get(ctx, key)
	if errors.Is(cmd.Err(), redis.Nil) {
		return "", urlshort.ErrMissingKey
	}
	if cmd.Err() != nil {
		return "", cmd.Err()
	}
	return cmd.Val(), nil
}
