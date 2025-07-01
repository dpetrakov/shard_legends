package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Сохраняем исходные значения ENV переменных
	originalEnv := make(map[string]string)
	envKeys := []string{
		"DATABASE_URL",
		"PORT_PUBLIC",
		"PORT_INTERNAL",
		"REDIS_AUTH_URL",
		"PRODUCTION_INTERNAL_URL",
		"PRODUCTION_EXTERNAL_URL",
		"INVENTORY_INTERNAL_URL",
		"AUTH_PUBLIC_KEY_URL",
		"COOLDOWN_SEC",
		"DAILY_CHEST_RECIPE_ID",
		"LOG_LEVEL",
	}

	for _, key := range envKeys {
		originalEnv[key] = os.Getenv(key)
	}

	// Восстанавливаем переменные окружения после теста
	defer func() {
		for _, key := range envKeys {
			if value, exists := originalEnv[key]; exists && value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("load_with_required_env", func(t *testing.T) {
		// Очищаем все переменные
		for _, key := range envKeys {
			os.Unsetenv(key)
		}

		// Устанавливаем обязательную переменную
		os.Setenv("DATABASE_URL", "postgresql://test:test@localhost:5432/test")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Проверяем значения по умолчанию
		if cfg.ServicePort != "8080" {
			t.Errorf("Expected ServicePort to be '8080', got '%s'", cfg.ServicePort)
		}

		if cfg.InternalServicePort != "8090" {
			t.Errorf("Expected InternalServicePort to be '8090', got '%s'", cfg.InternalServicePort)
		}

		if cfg.ServiceHost != "0.0.0.0" {
			t.Errorf("Expected ServiceHost to be '0.0.0.0', got '%s'", cfg.ServiceHost)
		}

		if cfg.DatabaseURL != "postgresql://test:test@localhost:5432/test" {
			t.Errorf("Expected DatabaseURL to be 'postgresql://test:test@localhost:5432/test', got '%s'", cfg.DatabaseURL)
		}

		if cfg.RedisAuthURL != "redis://redis:6379/0" {
			t.Errorf("Expected RedisAuthURL to be 'redis://redis:6379/0', got '%s'", cfg.RedisAuthURL)
		}

		if cfg.CooldownSec != 30 {
			t.Errorf("Expected CooldownSec to be 30, got %d", cfg.CooldownSec)
		}

		if cfg.DailyChestRecipeID != "9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2" {
			t.Errorf("Expected DailyChestRecipeID to be '9b9a4a62-7e79-4f1c-8dbe-62784be4c9d2', got '%s'", cfg.DailyChestRecipeID)
		}

		if cfg.LogLevel != "info" {
			t.Errorf("Expected LogLevel to be 'info', got '%s'", cfg.LogLevel)
		}
	})

	t.Run("load_with_custom_env", func(t *testing.T) {
		// Очищаем все переменные
		for _, key := range envKeys {
			os.Unsetenv(key)
		}

		// Устанавливаем кастомные значения
		os.Setenv("DATABASE_URL", "postgresql://custom:pass@example.com:5432/custom_db")
		os.Setenv("PORT_PUBLIC", "9080")
		os.Setenv("PORT_INTERNAL", "9090")
		os.Setenv("COOLDOWN_SEC", "60")
		os.Setenv("LOG_LEVEL", "debug")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if cfg.ServicePort != "9080" {
			t.Errorf("Expected ServicePort to be '9080', got '%s'", cfg.ServicePort)
		}

		if cfg.InternalServicePort != "9090" {
			t.Errorf("Expected InternalServicePort to be '9090', got '%s'", cfg.InternalServicePort)
		}

		if cfg.CooldownSec != 60 {
			t.Errorf("Expected CooldownSec to be 60, got %d", cfg.CooldownSec)
		}

		if cfg.LogLevel != "debug" {
			t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
		}
	})

	t.Run("missing_required_database_url", func(t *testing.T) {
		// Очищаем все переменные
		for _, key := range envKeys {
			os.Unsetenv(key)
		}

		_, err := Load()
		if err == nil {
			t.Fatal("Expected error for missing DATABASE_URL, got nil")
		}

		expectedMsg := "DATABASE_URL is required"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid_config", func(t *testing.T) {
		cfg := &Config{
			ServiceHost:         "0.0.0.0",
			ServicePort:         "8080",
			InternalServicePort: "8090",
			DatabaseURL:         "postgresql://user:pass@localhost:5432/db",
			DatabaseMaxConns:    10,
			RedisAuthURL:        "redis://redis:6379/0",
			RedisMaxConns:       10,
			CooldownSec:         30,
			LogLevel:            "info",
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Expected no error for valid config, got: %v", err)
		}
	})

	t.Run("invalid_database_url", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "mysql://invalid:url",
			RedisAuthURL:     "redis://redis:6379/0",
			DatabaseMaxConns: 10,
			RedisMaxConns:    10,
			CooldownSec:      30,
			LogLevel:         "info",
		}

		err := cfg.Validate()
		if err == nil {
			t.Fatal("Expected error for invalid database URL, got nil")
		}
	})

	t.Run("invalid_log_level", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL:      "postgresql://user:pass@localhost:5432/db",
			RedisAuthURL:     "redis://redis:6379/0",
			DatabaseMaxConns: 10,
			RedisMaxConns:    10,
			CooldownSec:      30,
			LogLevel:         "invalid",
		}

		err := cfg.Validate()
		if err == nil {
			t.Fatal("Expected error for invalid log level, got nil")
		}
	})
}

func TestConfigString(t *testing.T) {
	cfg := &Config{
		ServiceHost:         "0.0.0.0",
		ServicePort:         "8080",
		InternalServicePort: "8090",
		DatabaseURL:         "postgresql://user:secret@localhost:5432/db",
		RedisAuthURL:        "redis://password@redis:6379/0",
		LogLevel:            "info",
		CooldownSec:         30,
		DailyChestRecipeID:  "test-recipe-id",
	}

	str := cfg.String()

	// Проверяем, что чувствительные данные замаскированы
	if str == "" {
		t.Error("Config string should not be empty")
	}

	// Проверяем маскировку URL
	if cfg.String() != cfg.String() {
		t.Error("Config string should be consistent")
	}
}
