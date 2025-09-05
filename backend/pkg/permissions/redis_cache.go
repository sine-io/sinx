package permissions

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisUserPermCache 基于 Redis 的权限缓存
type RedisUserPermCache struct {
	cli    *redis.Client
	ttl    time.Duration
	prefix string
}

func NewRedisUserPermCache(cli *redis.Client, ttl time.Duration) *RedisUserPermCache {
	return &RedisUserPermCache{cli: cli, ttl: ttl, prefix: "user_perms:"}
}

func (c *RedisUserPermCache) key(userID uint) string {
	return c.prefix + strconv.FormatUint(uint64(userID), 10)
}

// Get 返回权限集合；若无或解析失败返回 nil
func (c *RedisUserPermCache) Get(ctx context.Context, userID uint) (map[string]struct{}, error) {
	if c.cli == nil {
		return nil, nil
	}
	val, err := c.cli.Get(ctx, c.key(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var arr []string
	if err := json.Unmarshal([]byte(val), &arr); err != nil {
		return nil, err
	}
	m := make(map[string]struct{}, len(arr))
	for _, p := range arr {
		m[p] = struct{}{}
	}
	return m, nil
}

// Set 写入权限集合
func (c *RedisUserPermCache) Set(ctx context.Context, userID uint, perms map[string]struct{}) error {
	if c.cli == nil {
		return nil
	}
	arr := make([]string, 0, len(perms))
	for p := range perms {
		arr = append(arr, p)
	}
	b, _ := json.Marshal(arr)
	return c.cli.Set(ctx, c.key(userID), b, c.ttl).Err()
}

// Invalidate 删除指定用户权限缓存
func (c *RedisUserPermCache) Invalidate(ctx context.Context, userID uint) {
	if c.cli == nil {
		return
	}
	_ = c.cli.Del(ctx, c.key(userID)).Err()
}

// InvalidateUsers 批量删除
func (c *RedisUserPermCache) InvalidateUsers(ctx context.Context, userIDs []uint) {
	if c.cli == nil || len(userIDs) == 0 {
		return
	}
	keys := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		keys = append(keys, c.key(id))
	}
	_ = c.cli.Del(ctx, keys...).Err()
}
