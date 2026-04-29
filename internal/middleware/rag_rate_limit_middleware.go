package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	windowStart time.Time
	count       int
}

var (
	ragRateLimitStore sync.Map
	ragStoreMu        sync.Mutex
)

func RAGQueryRateLimitMiddleware() gin.HandlerFunc {
	maxRequests := readIntEnv("RAG_QUERY_RATE_LIMIT", 30)
	windowSecs := readIntEnv("RAG_QUERY_WINDOW_SECONDS", 60)
	window := time.Duration(windowSecs) * time.Second

	if maxRequests <= 0 {
		maxRequests = 30
	}
	if window <= 0 {
		window = time.Minute
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now().UTC()

		ragStoreMu.Lock()
		value, _ := ragRateLimitStore.LoadOrStore(clientIP, rateLimitEntry{windowStart: now, count: 0})
		entry := value.(rateLimitEntry)

		if now.Sub(entry.windowStart) >= window {
			entry = rateLimitEntry{windowStart: now, count: 0}
		}

		if entry.count >= maxRequests {
			ragStoreMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please retry shortly",
			})
			c.Abort()
			return
		}

		entry.count++
		ragRateLimitStore.Store(clientIP, entry)
		ragStoreMu.Unlock()

		c.Next()
	}
}

func readIntEnv(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
