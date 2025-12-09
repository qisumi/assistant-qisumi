package config

import (
	"os"
	"strconv"
)

// Config 应用程序配置
type Config struct {
	DB     DBConfig
	HTTP   HTTPConfig
	JWT    JWTConfig
	Crypto CryptoConfig
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// HTTPConfig HTTP服务器配置
type HTTPConfig struct {
	Host string
	Port string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	ExpireHour int
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	APIKeyEncryptionKey string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	expireHour, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOUR", "24"))

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			Username: getEnv("DB_USERNAME", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_DATABASE", "assistant_qisumi"),
		},
		HTTP: HTTPConfig{
			Host: getEnv("HTTP_HOST", "0.0.0.0"),
			Port: getEnv("HTTP_PORT", "8080"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			ExpireHour: expireHour,
		},
		Crypto: CryptoConfig{
			APIKeyEncryptionKey: getEnv("API_KEY_ENCRYPTION_KEY", "your-32-byte-encryption-key"),
		},
	}, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
