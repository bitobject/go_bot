package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware создает структурированное логирование
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Структурированное логирование
		logrus.WithFields(logrus.Fields{
			"timestamp": param.TimeStamp.Format(time.RFC3339),
			"status":    param.StatusCode,
			"latency":   param.Latency,
			"client_ip": param.ClientIP,
			"method":    param.Method,
			"path":      param.Path,
			"user_agent": param.Request.UserAgent(),
		}).Info("HTTP Request")

		return ""
	})
}

// RequestLogger логирует детали запроса
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log details
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		logrus.WithFields(logrus.Fields{
			"method":     method,
			"path":       path,
			"status":     status,
			"latency":    latency,
			"client_ip":  clientIP,
			"user_agent": c.Request.UserAgent(),
		}).Info("Request completed")
	}
} 