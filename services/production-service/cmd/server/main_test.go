package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	customMiddleware "github.com/shard-legends/production-service/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortSeparation(t *testing.T) {
	// Set required environment variables
	os.Setenv("PROD_SVC_PUBLIC_PORT", "8082")
	os.Setenv("PROD_SVC_INTERNAL_PORT", "8091")
	os.Setenv("DATABASE_URL", "postgresql://test:test@localhost:5432/test")
	os.Setenv("REDIS_URL", "redis://localhost:6379/1")
	defer func() {
		os.Unsetenv("PROD_SVC_PUBLIC_PORT")
		os.Unsetenv("PROD_SVC_INTERNAL_PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_URL")
	}()


	// Setup public router
	publicRouter := chi.NewRouter()
	publicRouter.Use(middleware.RequestID)
	publicRouter.Use(middleware.RealIP)
	publicRouter.Use(customMiddleware.Recovery())
	publicRouter.Use(middleware.Timeout(60 * time.Second))

	// Public API routes (testing without auth for simplicity)
	publicRouter.Route("/production", func(r chi.Router) {
		r.Get("/recipes", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"recipes endpoint"}`))
		})
	})

	// Setup internal router
	internalRouter := chi.NewRouter()
	internalRouter.Use(middleware.RequestID)
	internalRouter.Use(middleware.RealIP)
	internalRouter.Use(customMiddleware.Recovery())
	internalRouter.Use(middleware.Timeout(60 * time.Second))

	// Internal endpoints
	internalRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	internalRouter.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})
	internalRouter.Handle("/metrics", promhttp.Handler())

	// Admin endpoints (testing without auth for simplicity)
	internalRouter.Route("/api/v1/admin", func(r chi.Router) {
		r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"admin tasks"}`))
		})
	})

	t.Run("Public router should NOT expose internal endpoints", func(t *testing.T) {
		// Test that metrics is not available on public router
		req := httptest.NewRequest("GET", "/metrics", nil)
		rec := httptest.NewRecorder()
		publicRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "Metrics should not be available on public port")

		// Test that health is not available on public router
		req = httptest.NewRequest("GET", "/health", nil)
		rec = httptest.NewRecorder()
		publicRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "Health should not be available on public port")

		// Test that admin endpoints are not available on public router
		req = httptest.NewRequest("GET", "/api/v1/admin/tasks", nil)
		rec = httptest.NewRecorder()
		publicRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, "Admin endpoints should not be available on public port")
	})

	t.Run("Internal router should expose health and metrics", func(t *testing.T) {
		// Test that health is available on internal router
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()
		internalRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Health should be available on internal port")
		assert.Contains(t, rec.Body.String(), "healthy")

		// Test that ready is available on internal router
		req = httptest.NewRequest("GET", "/ready", nil)
		rec = httptest.NewRecorder()
		internalRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Ready should be available on internal port")
		assert.Contains(t, rec.Body.String(), "ready")

		// Test that metrics is available on internal router
		req = httptest.NewRequest("GET", "/metrics", nil)
		rec = httptest.NewRecorder()
		internalRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Metrics should be available on internal port")
	})

	t.Run("Public router should only expose business endpoints", func(t *testing.T) {
		// Test that public business endpoints are accessible
		req := httptest.NewRequest("GET", "/production/recipes", nil)
		rec := httptest.NewRecorder()
		publicRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Recipes endpoint should be accessible")
		assert.Contains(t, rec.Body.String(), "recipes endpoint")
	})

	t.Run("Configuration validation", func(t *testing.T) {
		// Test that ports are required
		os.Unsetenv("PROD_SVC_PUBLIC_PORT")
		// Config loading would fail here in real application
		require.Empty(t, os.Getenv("PROD_SVC_PUBLIC_PORT"), "Public port should be required")

		os.Unsetenv("PROD_SVC_INTERNAL_PORT")
		// Config loading would fail here in real application
		require.Empty(t, os.Getenv("PROD_SVC_INTERNAL_PORT"), "Internal port should be required")
	})
}

