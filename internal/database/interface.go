package database

import (
	"context"
)

type DB interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)

	Ping(ctx context.Context) error
	Close() error
}
