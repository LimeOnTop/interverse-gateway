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

func (b *AccessBlacklist) CurrentVersion(ctx context.Context, userID string) (int64, error) {
	if b == nil || b.client == nil {
		return 0, fmt.Errorf("session store unavailable")
	}
	v, err := b.client.Get(ctx, "auth:version:"+userID).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return v, err
}
