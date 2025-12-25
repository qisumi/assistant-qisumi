package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用程序配置
type Config struct {
	DB     DBConfig
	HTTP   HTTPConfig
	JWT    JWTConfig
	Crypto CryptoConfig
	LLM    LLMConfig
	Log    LogConfig
}

// LLMConfig LLM配置
type LLMConfig struct {
	APIKey          string
	ModelName       string
	APIBaseURL      string
	ThinkingType    string // disabled, enabled, auto
	ReasoningEffort string // low, medium, high, minimal
}

// DBConfig 数据库配置
type DBConfig struct {
	Type     string // mysql or sqlite
	Host     string
	Port     string
	Username string
	Password string
	Database string
	FilePath string // for sqlite
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

// LogConfig 日志配置
type LogConfig struct {
	Level string // debug, info, warn, error
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	// 加载.env文件，忽略不存在的错误
	_ = godotenv.Load()

	// 获取可执行文件所在目录
	execDir, err := GetExecutableDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable directory: %w", err)
	}

	expireHour, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOUR", "24"))

	// 默认数据库文件路径为可执行文件所在目录
	defaultDBPath := filepath.Join(execDir, "assistant.db")

	return &Config{
		DB: DBConfig{
			Type:     getEnv("DB_TYPE", "mysql"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			Username: getEnv("DB_USERNAME", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_DATABASE", "assistant_qisumi"),
			FilePath: getEnv("DB_FILE_PATH", defaultDBPath),
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
			APIKey:          getEnv("LLM_API_KEY", ""),
			ModelName:       getEnv("LLM_MODEL_NAME", "qwen-plus"),
			APIBaseURL:      getEnv("LLM_API_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
			ThinkingType:    getEnv("LLM_THINKING_TYPE", "auto"),
			ReasoningEffort: getEnv("LLM_REASONING_EFFORT", "medium"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
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

// GetExecutableDir 获取可执行文件所在目录
// 兼容 go run 和二进制运行两种模式
func GetExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// 解析符号链接（如果有）
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", err
	}

	// 检测是否是 go run 模式
	// go run 会创建临时目录，路径中通常包含 "go-build" 或 "Temp"
	// 如果是临时目录，则返回当前工作目录
	if isGoRunMode(execPath) {
		// go run 模式，返回当前工作目录
		return os.Getwd()
	}

	// 二进制运行模式，返回可执行文件所在目录
	return filepath.Dir(execPath), nil
}

// isGoRunMode 检测是否是 go run 模式
func isGoRunMode(execPath string) bool {
	// Windows 临时目录特征
	if filepath.Base(filepath.Dir(filepath.Dir(execPath))) == "Temp" {
		return true
	}
	// Linux/Mac 临时目录特征
	if filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(execPath)))) == "tmp" {
		return true
	}
	// go build 临时目录特征
	if strings.Contains(execPath, "go-build") {
		return true
	}
	return false
}
