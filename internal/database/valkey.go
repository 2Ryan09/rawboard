package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type ValkeyDB struct {
	client *redis.Client
}

func NewValkeyDB() (*ValkeyDB, error) {
	// Get connection URI from environment - try multiple common environment variables
	uri := os.Getenv("VALKEY_URI")
	if uri == "" {
		uri = os.Getenv("REDIS_URL")
	}
	if uri == "" {
		uri = os.Getenv("DATABASE_URL")
	}
	if uri == "" {
		// Try building from VALKEY_URL or component parts
		if valkeyURL := os.Getenv("VALKEY_URL"); valkeyURL != "" {
			uri = "redis://" + valkeyURL
		} else if host := os.Getenv("REDIS_HOST"); host != "" {
			port := os.Getenv("REDIS_PORT")
			if port == "" {
				port = "6379"
			}
			uri = "redis://" + host + ":" + port
		} else {
			uri = "redis://localhost:6379"
		}
	}

	opts, err := redis.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Valkey URI: %w", err)
	}

	// Set reasonable timeouts for cloud deployments
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 5 * time.Second
	opts.WriteTimeout = 5 * time.Second

	client := redis.NewClient(opts)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Valkey: %w", err)
	}

	return &ValkeyDB{client: client}, nil
}

func (v *ValkeyDB) Set(ctx context.Context, key string, value interface{}) error {
	return v.client.Set(ctx, key, value, 0).Err() // 0 = no expiration
}

func (v *ValkeyDB) Get(ctx context.Context, key string) (string, error) {
	return v.client.Get(ctx, key).Result()
}

func (v *ValkeyDB) Ping(ctx context.Context) error {
	return v.client.Ping(ctx).Err()
}

func (v *ValkeyDB) Close() error {
	return v.client.Close()
}
