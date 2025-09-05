package permissions

import (
	"sync"
	"time"
)

// userPermCacheItem 缓存项
type userPermCacheItem struct {
	perms map[string]struct{}
	exp   time.Time
}

// UserPermCache 简单内存缓存（进程级，后续可替换为 Redis）
type UserPermCache struct {
	ttl   time.Duration
	mu    sync.RWMutex
	store map[uint]*userPermCacheItem
}

func NewUserPermCache(ttl time.Duration) *UserPermCache {
	return &UserPermCache{ttl: ttl, store: make(map[uint]*userPermCacheItem)}
}

// Get 返回缓存的权限集合，若不存在或过期返回 nil
func (c *UserPermCache) Get(userID uint) map[string]struct{} {
	c.mu.RLock()
	item, ok := c.store[userID]
	c.mu.RUnlock()
	if !ok || time.Now().After(item.exp) {
		if ok { // 过期清理
			c.mu.Lock()
			delete(c.store, userID)
			c.mu.Unlock()
		}
		return nil
	}
	return item.perms
}

// Set 写入权限集合
func (c *UserPermCache) Set(userID uint, perms map[string]struct{}) {
	c.mu.Lock()
	c.store[userID] = &userPermCacheItem{perms: perms, exp: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// Invalidate 使用户权限缓存失效
func (c *UserPermCache) Invalidate(userID uint) {
	c.mu.Lock()
	delete(c.store, userID)
	c.mu.Unlock()
}

// InvalidateUsers 批量失效
func (c *UserPermCache) InvalidateUsers(userIDs []uint) {
	c.mu.Lock()
	for _, id := range userIDs {
		delete(c.store, id)
	}
	c.mu.Unlock()
}
