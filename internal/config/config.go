package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用程序配置
type Config struct {
	DB     DBConfig
	HTTP   HTTPConfig
	JWT    JWTConfig
	Crypto CryptoConfig
	LLM    LLMConfig
}

// LLMConfig LLM配置
type LLMConfig struct {
	APIKey     string
	ModelName  string
	APIBaseURL string
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
	// 加载.env文件，忽略不存在的错误
	_ = godotenv.Load()
	
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
			Port: getEnv("HTTP_PORT", "4569"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			ExpireHour: expireHour,
		},
		Crypto: CryptoConfig{
			APIKeyEncryptionKey: getEnv("API_KEY_ENCRYPTION_KEY", "your-32-byte-encryption-key"),
		},
		LLM: LLMConfig{
			APIKey:     getEnv("LLM_API_KEY", ""),
			ModelName:  getEnv("LLM_MODEL_NAME", "qwen-plus"),
			APIBaseURL: getEnv("LLM_API_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
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
