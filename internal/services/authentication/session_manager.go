package authentication

import (
	"context"
	"time"
)

type sessionManager interface {
	Load(ctx context.Context, token string) (context.Context, error)
	RenewToken(ctx context.Context) error
	Get(ctx context.Context, key string) interface{}
	Put(ctx context.Context, key string, val interface{})
	Commit(ctx context.Context) (string, time.Time, error)
	Destroy(ctx context.Context) error
}
