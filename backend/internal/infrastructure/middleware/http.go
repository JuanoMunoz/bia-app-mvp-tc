package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func JSONMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Header("Content-Type", "application/json; charset=utf-8")
		context.Next()
	}
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		startedAt := time.Now()
		context.Next()
		logger.Info("http request", "method", context.Request.Method, "path", context.FullPath(), "status", context.Writer.Status(), "duration", time.Since(startedAt))
	}
}

func LocalCORSMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		origin := context.GetHeader("Origin")
		if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
			context.Header("Access-Control-Allow-Origin", origin)
			context.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			context.Header("Access-Control-Allow-Headers", "Content-Type")
		}
		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(204)
			return
		}
		context.Next()
	}
}
