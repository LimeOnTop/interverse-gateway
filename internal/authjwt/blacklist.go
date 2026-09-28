package authjwt

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// AccessBlacklist checks logout revoke keys written by auth-service (DB 1).
type AccessBlacklist struct {
	client *redis.Client
}

func NewAccessBlacklist(client *redis.Client) *AccessBlacklist {
	return &AccessBlacklist{client: client}
}

func accessBLKey(jti string) string {
	return "auth:bl:access:" + jti
}

func (b *AccessBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if b == nil || b.client == nil {
		return false, nil
	}
	n, err := b.client.Exists(ctx, accessBLKey(jti)).Result()
	if err != nil {
		return false, fmt.Errorf("check access blacklist: %w", err)
	}
	return n > 0, nil
}
