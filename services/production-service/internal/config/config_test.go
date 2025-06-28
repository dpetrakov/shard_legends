package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to clear environment variables
func clearEnv(t *testing.T) {
	envVars := []string{
		"PROD_SVC_DATABASE_URL",
		"PROD_SVC_REDIS_URL",
		"PROD_SVC_REDIS_AUTH_URL",
		"PROD_SVC_SERVER_PORT",
		"PROD_SVC_SERVER_INTERNAL_PORT",
		"PROD_SVC_AUTH_PUBLIC_KEY_URL",
		"PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_BASE_URL",
		"PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_BASE_URL",
		"PROD_SVC_SERVER_HOST",
		"PROD_SVC_LOGGING_LEVEL",
		"PROD_SVC_DATABASE_MAX_CONNECTIONS",
		"PROD_SVC_REDIS_MAX_CONNECTIONS",
		"PROD_SVC_SERVER_READ_TIMEOUT",
		"PROD_SVC_SERVER_WRITE_TIMEOUT",
		"PROD_SVC_DATABASE_PING_TIMEOUT",
		"PROD_SVC_AUTH_CACHE_TTL",
		"PROD_SVC_REDIS_MAX_RETRIES",
	}
	
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

// Helper function to set required environment variables
func setRequiredEnv(t *testing.T) {
	os.Setenv("PROD_SVC_DATABASE_URL", "postgres://user:pass@localhost:5432/test")
	os.Setenv("PROD_SVC_REDIS_URL", "redis://localhost:6379/1")
	os.Setenv("PROD_SVC_REDIS_AUTH_URL", "redis://localhost:6379/0")
	os.Setenv("PROD_SVC_SERVER_PORT", "8080")
	os.Setenv("PROD_SVC_SERVER_INTERNAL_PORT", "8081")
	os.Setenv("PROD_SVC_AUTH_PUBLIC_KEY_URL", "http://auth:8080/key.pem")
	os.Setenv("PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_BASE_URL", "http://inventory:8080")
	os.Setenv("PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_BASE_URL", "http://user:8080")
}

func TestConfig_Load_Success(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	setRequiredEnv(t)
	
	cfg, err := Load()
	
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	
	// Test required fields are set
	assert.Equal(t, "postgres://user:pass@localhost:5432/test", cfg.Database.URL)
	assert.Equal(t, "redis://localhost:6379/1", cfg.Redis.URL)
	assert.Equal(t, "redis://localhost:6379/0", cfg.Redis.AuthURL)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "8081", cfg.Server.InternalPort)
	assert.Equal(t, "http://auth:8080/key.pem", cfg.Auth.PublicKeyURL)
	assert.Equal(t, "http://inventory:8080", cfg.ExternalServices.InventoryService.BaseURL)
	assert.Equal(t, "http://user:8080", cfg.ExternalServices.UserService.BaseURL)
	
	// Test defaults are applied
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 25, cfg.Database.MaxConnections)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestConfig_Load_WithOptionalOverrides(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	setRequiredEnv(t)
	
	// Set optional overrides
	os.Setenv("PROD_SVC_SERVER_HOST", "127.0.0.1")
	os.Setenv("PROD_SVC_LOGGING_LEVEL", "debug")
	os.Setenv("PROD_SVC_DATABASE_MAX_CONNECTIONS", "50")
	os.Setenv("PROD_SVC_REDIS_MAX_CONNECTIONS", "20")
	os.Setenv("PROD_SVC_SERVER_READ_TIMEOUT", "30s")
	
	cfg, err := Load()
	
	require.NoError(t, err)
	
	// Test overrides are applied
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, 50, cfg.Database.MaxConnections)
	assert.Equal(t, 20, cfg.Redis.MaxConnections)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
}

func TestConfig_Validate_MissingRequiredFields(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	
	// Test that validation fails fast when any required field is missing
	_, err := Load()
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration field")
	assert.Contains(t, err.Error(), "is not set")
}

func TestConfig_ValidationLogic(t *testing.T) {
	t.Run("InvalidZeroTimeout", func(t *testing.T) {
		clearEnv(t)
		setRequiredEnv(t)
		
		os.Setenv("PROD_SVC_SERVER_READ_TIMEOUT", "0s")
		_, err := Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})
	
	t.Run("TimeoutTooLarge", func(t *testing.T) {
		clearEnv(t)
		setRequiredEnv(t)
		
		os.Setenv("PROD_SVC_SERVER_READ_TIMEOUT", "15m")
		_, err := Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "seems too large")
	})
	
	t.Run("InvalidDatabaseConnections", func(t *testing.T) {
		clearEnv(t)
		setRequiredEnv(t)
		
		os.Setenv("PROD_SVC_DATABASE_MAX_CONNECTIONS", "0")
		_, err := Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database.max_connections must be positive")
	})
	
	t.Run("InvalidRedisConnections", func(t *testing.T) {
		clearEnv(t)
		setRequiredEnv(t)
		
		os.Setenv("PROD_SVC_REDIS_MAX_CONNECTIONS", "-1")
		_, err := Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis.max_connections must be positive")
	})
	
	t.Run("InvalidRedisRetries", func(t *testing.T) {
		clearEnv(t)
		setRequiredEnv(t)
		
		os.Setenv("PROD_SVC_REDIS_MAX_RETRIES", "-1")
		_, err := Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redis.max_retries cannot be negative")
	})
}

func TestConfig_GetCleanupConfig(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	setRequiredEnv(t)
	
	cfg, err := Load()
	require.NoError(t, err)
	
	cleanupConfig := cfg.GetCleanupConfig()
	
	assert.Equal(t, cfg.Cleanup.OrphanedTaskTimeout, cleanupConfig.OrphanedTaskTimeout)
	assert.Equal(t, cfg.Cleanup.CleanupInterval, cleanupConfig.CleanupInterval)
	assert.Equal(t, cfg.Cleanup.CleanupTimeout, cleanupConfig.CleanupTimeout)
}

func TestConfig_TimeoutParsing(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	setRequiredEnv(t)
	
	// Test various timeout formats
	os.Setenv("PROD_SVC_SERVER_READ_TIMEOUT", "30s")
	os.Setenv("PROD_SVC_SERVER_WRITE_TIMEOUT", "2m")
	os.Setenv("PROD_SVC_DATABASE_PING_TIMEOUT", "500ms")
	os.Setenv("PROD_SVC_AUTH_CACHE_TTL", "2h")
	
	cfg, err := Load()
	require.NoError(t, err)
	
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 2*time.Minute, cfg.Server.WriteTimeout)
	assert.Equal(t, 500*time.Millisecond, cfg.Database.PingTimeout)
	assert.Equal(t, 2*time.Hour, cfg.Auth.CacheTTL)
}

func TestConfig_FailFast_Behavior(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	
	// Test that the application fails fast when critical config is missing
	_, err := Load()
	
	require.Error(t, err)
	// Should fail quickly without trying to connect to services
	assert.Contains(t, err.Error(), "required configuration field")
}

func TestConfig_AllRequiredFieldsValidation(t *testing.T) {
	// Test that validation fails when any required field is missing
	// Since validation is fail-fast, we only check that an error occurs
	requiredFields := []string{
		"PROD_SVC_DATABASE_URL",
		"PROD_SVC_REDIS_URL",
		"PROD_SVC_REDIS_AUTH_URL",
		"PROD_SVC_SERVER_PORT",
		"PROD_SVC_SERVER_INTERNAL_PORT",
		"PROD_SVC_AUTH_PUBLIC_KEY_URL",
		"PROD_SVC_EXTERNAL_SERVICES_INVENTORY_SERVICE_BASE_URL",
		"PROD_SVC_EXTERNAL_SERVICES_USER_SERVICE_BASE_URL",
	}
	
	for _, envVar := range requiredFields {
		t.Run("Missing_"+envVar, func(t *testing.T) {
			clearEnv(t)
			setRequiredEnv(t)
			
			// Remove specific required field
			os.Unsetenv(envVar)
			
			_, err := Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "required configuration field")
			assert.Contains(t, err.Error(), "is not set")
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)
	setRequiredEnv(t)
	
	cfg, err := Load()
	require.NoError(t, err)
	
	// Test all default values are as expected
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 15*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)
	
	assert.Equal(t, 25, cfg.Database.MaxConnections)
	assert.Equal(t, 5*time.Minute, cfg.Database.MaxIdleTime)
	assert.Equal(t, 1*time.Minute, cfg.Database.HealthCheckPeriod)
	assert.Equal(t, 5*time.Second, cfg.Database.PingTimeout)
	
	assert.Equal(t, 10, cfg.Redis.MaxConnections)
	assert.Equal(t, 3*time.Second, cfg.Redis.ReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.Redis.WriteTimeout)
	assert.Equal(t, 3, cfg.Redis.MaxRetries)
	assert.Equal(t, 5*time.Second, cfg.Redis.PingTimeout)
	
	assert.Equal(t, 1*time.Hour, cfg.Auth.CacheTTL)
	assert.Equal(t, 24*time.Hour, cfg.Auth.RefreshInterval)
	
	assert.Equal(t, "info", cfg.Logging.Level)
	
	assert.Equal(t, 10*time.Second, cfg.ExternalServices.InventoryService.Timeout)
	assert.Equal(t, 5*time.Second, cfg.ExternalServices.UserService.Timeout)
	
	assert.Equal(t, 60*time.Second, cfg.Timeouts.HTTPMiddleware)
	assert.Equal(t, 10*time.Second, cfg.Timeouts.JWTValidatorClient)
	assert.Equal(t, 30*time.Second, cfg.Timeouts.GracefulShutdown)
	assert.Equal(t, 2*time.Second, cfg.Timeouts.DatabaseHealth)
	assert.Equal(t, 2*time.Second, cfg.Timeouts.RedisHealth)
	
	assert.Equal(t, 5*time.Minute, cfg.Cleanup.OrphanedTaskTimeout)
	assert.Equal(t, 5*time.Minute, cfg.Cleanup.CleanupInterval)
	assert.Equal(t, 5*time.Minute, cfg.Cleanup.CleanupTimeout)
	
	assert.Equal(t, 10*time.Second, cfg.Metrics.UpdateInterval)
}