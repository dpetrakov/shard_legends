package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnvVars := make(map[string]string)
	envVars := []string{
		"INVENTORY_SERVICE_HOST", "INVENTORY_SERVICE_PORT",
		"DATABASE_URL", "DATABASE_MAX_CONNECTIONS",
		"REDIS_URL", "REDIS_MAX_CONNECTIONS",
		"LOG_LEVEL", "METRICS_PORT",
	}
	
	for _, key := range envVars {
		originalEnvVars[key] = os.Getenv(key)
	}
	
	// Clean up function
	cleanup := func() {
		for _, key := range envVars {
			if original, exists := originalEnvVars[key]; exists && original != "" {
				os.Setenv(key, original)
			} else {
				os.Unsetenv(key)
			}
		}
	}
	defer cleanup()

	t.Run("success with all env vars", func(t *testing.T) {
		// Clean env
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		
		// Set test values
		os.Setenv("INVENTORY_SERVICE_HOST", "127.0.0.1")
		os.Setenv("INVENTORY_SERVICE_PORT", "8081")
		os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
		os.Setenv("DATABASE_MAX_CONNECTIONS", "20")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Setenv("REDIS_MAX_CONNECTIONS", "15")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("METRICS_PORT", "9091")

		cfg, err := Load()
		require.NoError(t, err)
		
		assert.Equal(t, "127.0.0.1", cfg.ServiceHost)
		assert.Equal(t, "8081", cfg.ServicePort)
		assert.Equal(t, "postgresql://user:pass@localhost:5432/testdb", cfg.DatabaseURL)
		assert.Equal(t, 20, cfg.DatabaseMaxConns)
		assert.Equal(t, "redis://localhost:6379", cfg.RedisURL)
		assert.Equal(t, 15, cfg.RedisMaxConns)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, "9091", cfg.MetricsPort)
	})

	t.Run("success with defaults", func(t *testing.T) {
		// Clean env
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		
		// Set only required values
		os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
		os.Setenv("REDIS_URL", "redis://localhost:6379")

		cfg, err := Load()
		require.NoError(t, err)
		
		// Check defaults
		assert.Equal(t, "0.0.0.0", cfg.ServiceHost)
		assert.Equal(t, "8080", cfg.ServicePort)
		assert.Equal(t, 10, cfg.DatabaseMaxConns)
		assert.Equal(t, 10, cfg.RedisMaxConns)
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, "9090", cfg.MetricsPort)
	})

	t.Run("missing DATABASE_URL", func(t *testing.T) {
		// Clean env
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		
		os.Setenv("REDIS_URL", "redis://localhost:6379")

		cfg, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL is required")
		assert.Nil(t, cfg)
	})

	t.Run("missing REDIS_URL", func(t *testing.T) {
		// Clean env
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		
		os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")

		cfg, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "REDIS_URL is required")
		assert.Nil(t, cfg)
	})

	t.Run("invalid DATABASE_MAX_CONNECTIONS", func(t *testing.T) {
		// Clean env
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		
		os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Setenv("DATABASE_MAX_CONNECTIONS", "not_a_number")

		cfg, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid DATABASE_MAX_CONNECTIONS")
		assert.Nil(t, cfg)
	})

	t.Run("invalid REDIS_MAX_CONNECTIONS", func(t *testing.T) {
		// Clean env
		for _, key := range envVars {
			os.Unsetenv(key)
		}
		
		os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Setenv("REDIS_MAX_CONNECTIONS", "invalid")

		cfg, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid REDIS_MAX_CONNECTIONS")
		assert.Nil(t, cfg)
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    10,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid postgres:// URL", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgres://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    10,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid database URL", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "mysql://user:pass@localhost:3306/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    10,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL must start with postgresql:// or postgres://")
	})

	t.Run("invalid redis URL", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "memcached://localhost:11211",
			RedisMaxConns:    10,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "REDIS_URL must start with redis://")
	})

	t.Run("database connections too low", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 0,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    10,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_MAX_CONNECTIONS must be between 1 and 100")
	})

	t.Run("database connections too high", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 101,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    10,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_MAX_CONNECTIONS must be between 1 and 100")
	})

	t.Run("redis connections too low", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    0,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "REDIS_MAX_CONNECTIONS must be between 1 and 100")
	})

	t.Run("redis connections too high", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    101,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "REDIS_MAX_CONNECTIONS must be between 1 and 100")
	})

	t.Run("invalid log level", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "redis://localhost:6379",
			RedisMaxConns:    10,
			LogLevel:         "invalid",
		}

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LOG_LEVEL must be one of: debug, info, warn, error")
	})

	t.Run("all valid log levels", func(t *testing.T) {
		validLevels := []string{"debug", "info", "warn", "error"}
		
		for _, level := range validLevels {
			cfg := &Config{
				DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
				DatabaseMaxConns: 10,
				RedisURL:         "redis://localhost:6379",
				RedisMaxConns:    10,
				LogLevel:         level,
			}

			err := cfg.Validate()
			assert.NoError(t, err, "log level %s should be valid", level)
		}
	})
}

func TestConfig_String(t *testing.T) {
	t.Run("string representation", func(t *testing.T) {
		cfg := &Config{
			ServiceHost:      "localhost",
			ServicePort:      "8080",
			DatabaseURL:      "postgresql://user:pass@localhost:5432/testdb",
			DatabaseMaxConns: 10,
			RedisURL:         "redis://user:pass@localhost:6379",
			RedisMaxConns:    10,
			LogLevel:         "info",
			MetricsPort:      "9090",
		}

		str := cfg.String()
		
		// Check that basic info is present
		assert.Contains(t, str, "localhost")
		assert.Contains(t, str, "8080")
		assert.Contains(t, str, "info")
		assert.Contains(t, str, "9090")
		
		// Check that URLs are masked (should contain ***)
		assert.Contains(t, str, "***")
	})
}

func TestMaskURL(t *testing.T) {
	t.Run("URL with credentials", func(t *testing.T) {
		url := "postgresql://user:password@localhost:5432/database"
		masked := maskURL(url)
		
		assert.Contains(t, masked, "postgresql://***@localhost:5432/database")
		assert.NotContains(t, masked, "password")
		assert.NotContains(t, masked, "user")
	})

	t.Run("Redis URL with credentials", func(t *testing.T) {
		url := "redis://user:secret@localhost:6379/0"
		masked := maskURL(url)
		
		assert.Contains(t, masked, "redis://***@localhost:6379/0")
		assert.NotContains(t, masked, "secret")
		assert.NotContains(t, masked, "user")
	})

	t.Run("URL without credentials", func(t *testing.T) {
		url := "redis://localhost:6379"
		masked := maskURL(url)
		
		// Should remain unchanged
		assert.Equal(t, url, masked)
	})

	t.Run("malformed URL", func(t *testing.T) {
		url := "not-a-url"
		masked := maskURL(url)
		
		// Should remain unchanged
		assert.Equal(t, url, masked)
	})

	t.Run("URL with @ but no credentials", func(t *testing.T) {
		url := "redis://@localhost:6379"
		
		// Should handle gracefully
		assert.NotPanics(t, func() {
			maskURL(url)
		})
	})
}