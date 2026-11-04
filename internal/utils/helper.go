package utils

import (
	"log"
	"member-shop-api/pkg/logger"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
)

func GetEnv(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intValue
	}

	return defaultValue
}

func NewLoggerWithPath(fileName string, level zerolog.Level) *zerolog.Logger {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ Unable to get working dir:", err)
	}

	path := filepath.Join(cwd, "/log", fileName)

	config := logger.LoggerConfig{
		Level:      level,
		Filename:   path,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     5,
		Compress:   true,
		IsDev:      GetEnv("APP_ENV", "development"),
	}
	return logger.NewLogger(config)
}

func MustGetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ Unable to get working dir:", err)
	}
	return dir
}

// ParseExpectedVersion trích xuất trường version từ payload update phục vụ Optimistic Concurrency Control (OCC)
func ParseExpectedVersion(updates map[string]any) *int64 {
	if v, exists := updates["version"]; exists {
		delete(updates, "version")
		switch val := v.(type) {
		case float64:
			vInt := int64(val)
			return &vInt
		case int64:
			return &val
		case int:
			vInt := int64(val)
			return &vInt
		}
	}
	return nil
}
