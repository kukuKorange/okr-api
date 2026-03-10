package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  int // hours
	RefreshTokenExpiry int // hours
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "goaltrack"),
			Password: getEnv("DB_PASSWORD", "goaltrack123"),
			DBName:   getEnv("DB_NAME", "goaltrack"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", "redis123"),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
			Bucket:    getEnv("MINIO_BUCKET", "goaltrack"),
			UseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "goaltrack-jwt-secret-change-me"),
			AccessTokenExpiry:  getEnvInt("JWT_ACCESS_EXPIRY", 2),
			RefreshTokenExpiry: getEnvInt("JWT_REFRESH_EXPIRY", 168), // 7 days
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
