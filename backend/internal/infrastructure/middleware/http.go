package middleware

import (
	"log/slog"
	"strings"
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
	return CORSMiddleware(nil)
}

// CORSMiddleware autoriza solo los orígenes de la lista. Cada entrada puede
// incluir un comodín "*" (ej. "https://*.vercel.app"). Sin lista, permite
// los orígenes de desarrollo local.
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}
	return func(context *gin.Context) {
		origin := context.GetHeader("Origin")
		if origin != "" && originAllowed(origin, allowedOrigins) {
			context.Header("Access-Control-Allow-Origin", origin)
			context.Header("Vary", "Origin")
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

func originAllowed(origin string, allowed []string) bool {
	for _, pattern := range allowed {
		if pattern == origin {
			return true
		}
		if strings.Contains(pattern, "*") {
			parts := strings.SplitN(pattern, "*", 2)
			if strings.HasPrefix(origin, parts[0]) && strings.HasSuffix(origin, parts[1]) {
				return true
			}
		}
	}
	return false
}
