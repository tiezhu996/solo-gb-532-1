package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditContext struct {
	RequestID, Method, Path string
	UserID                  uint
	Username, Role          string
	StartedAt               time.Time
}

func AuditContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		id, _ := userID.(uint)
		context := AuditContext{RequestID: c.GetString("request_id"), Method: c.Request.Method, Path: c.FullPath(), UserID: id, Username: c.GetString("username"), Role: c.GetString("role"), StartedAt: time.Now().UTC()}
		c.Set("audit_context", context)
		c.Next()
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.Header("X-Audit-Request-ID", context.RequestID)
		}
	}
}

type rateWindow struct {
	started time.Time
	count   int
}
type RateLimiter struct {
	mu       sync.Mutex
	capacity int
	window   time.Duration
	clients  map[string]rateWindow
}

func NewRateLimiter(capacity int, window time.Duration) *RateLimiter {
	return &RateLimiter{capacity: capacity, window: window, clients: make(map[string]rateWindow)}
}

func (l *RateLimiter) Middleware(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		key := scope + ":" + c.ClientIP()
		l.mu.Lock()
		entry := l.clients[key]
		if entry.started.IsZero() || now.Sub(entry.started) >= l.window {
			entry = rateWindow{started: now}
		}
		entry.count++
		l.clients[key] = entry
		if len(l.clients) > 2000 {
			for itemKey, item := range l.clients {
				if now.Sub(item.started) > l.window*2 {
					delete(l.clients, itemKey)
				}
			}
		}
		remaining := l.capacity - entry.count
		l.mu.Unlock()
		if remaining < 0 {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{"code": "RATE_LIMITED", "message": "请求过于频繁，请稍后重试"}, "request_id": c.GetString("request_id")})
			return
		}
		c.Next()
	}
}
