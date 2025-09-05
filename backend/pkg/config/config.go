package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	ListenAddr string
	LogLevel   string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret      string
	JWTExpireHours int
	JWTIssuer      string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

var cfg *Config

func LoadEnv() error {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		// .env文件不存在也不是错误，可能使用系统环境变量
	}

	cfg = &Config{
		AppEnv:     getEnv("APP_ENV", "development"),
		ListenAddr: getEnv("LISTEN_ADDR", ":8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "123456"),
		// 注意: 这里是业务数据库名称, 不应使用 Go Module 路径
		DBName:    getEnv("DB_NAME", "sinx"),
		DBSSLMode: getEnv("DB_SSL_MODE", "disable"),

		// JWT
		JWTSecret:      getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		JWTExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		JWTIssuer:      getEnv("JWT_ISSUER", "github.com/sine-io/sinx"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
	}

	return nil
}

func Get() *Config {
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
