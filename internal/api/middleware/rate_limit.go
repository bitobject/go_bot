package middleware

import (
	"net/http"
	"sync"
	"time"

	"go-bot/internal/config"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		rl.mutex.Lock()
		now := time.Now()
		
		// Очищаем старые запросы
		if requests, exists := rl.requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) <= rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[clientIP] = validRequests
		}
		
		// Проверяем лимит
		if len(rl.requests[clientIP]) >= rl.limit {
			rl.mutex.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}
		
		// Добавляем текущий запрос
		rl.requests[clientIP] = append(rl.requests[clientIP], now)
		rl.mutex.Unlock()
		
		c.Next()
	}
}

// RateLimitMiddleware создает rate limiter с настройками из конфигурации
func RateLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
	limiter := NewRateLimiter(
		cfg.RateLimitRequests,
		time.Duration(cfg.RateLimitWindowMinutes)*time.Minute,
	)
	return limiter.RateLimit()
} 