package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	startedAt time.Time
	count     int
}

type rateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateLimitEntry
	limit   int
	window  time.Duration
}

func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := &rateLimiter{entries: make(map[string]rateLimitEntry), limit: limit, window: window}
	return func(context *gin.Context) {
		if isHealthPath(context.FullPath()) {
			context.Next()
			return
		}
		key := context.ClientIP()
		now := time.Now()
		limiter.mu.Lock()
		entry := limiter.entries[key]
		if entry.startedAt.IsZero() || now.Sub(entry.startedAt) >= limiter.window {
			entry = rateLimitEntry{startedAt: now}
		}
		entry.count++
		allowed := entry.count <= limiter.limit
		remaining := max(0, limiter.limit-entry.count)
		resetAt := entry.startedAt.Add(limiter.window)
		limiter.entries[key] = entry
		limiter.mu.Unlock()

		context.Header("X-RateLimit-Limit", formatInt(limiter.limit))
		context.Header("X-RateLimit-Remaining", formatInt(remaining))
		context.Header("X-RateLimit-Reset", formatInt(int(resetAt.Unix())))
		if !allowed {
			context.Header("Retry-After", formatInt(max(1, int(time.Until(resetAt).Seconds()))))
			context.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		context.Next()
	}
}

func isHealthPath(path string) bool {
	return path == "/health" || path == "/isalive" || path == "/ping" || path == "/api/v1/health" || path == "/api/v1/isalive" || path == "/api/v1/ping"
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}
