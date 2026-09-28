package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewareAllowsConfiguredOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware([]string{"https://bia-app-mvp-tc.vercel.app", "https://*.vercel.app"}))
	router.GET("/ping", func(context *gin.Context) { context.String(http.StatusOK, "pong") })

	tests := []struct {
		origin  string
		allowed bool
	}{
		{origin: "https://bia-app-mvp-tc.vercel.app", allowed: true},
		{origin: "https://bia-app-mvp-xyz-juanomunozs-projects.vercel.app", allowed: true},
		{origin: "https://evil.com", allowed: false},
		{origin: "http://localhost:5173", allowed: false},
	}
	for _, test := range tests {
		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		request.Header.Set("Origin", test.origin)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		header := response.Header().Get("Access-Control-Allow-Origin")
		if test.allowed && header != test.origin {
			t.Errorf("origin %s: got header %q, want echoed origin", test.origin, header)
		}
		if !test.allowed && header != "" {
			t.Errorf("origin %s: got header %q, want empty", test.origin, header)
		}
	}
}

func TestLocalCORSMiddlewareKeepsDevOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LocalCORSMiddleware())
	router.GET("/ping", func(context *gin.Context) { context.String(http.StatusOK, "pong") })

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatal("localhost origin was not allowed")
	}
}
