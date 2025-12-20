package test

import (
	"assistant-qisumi/internal/config"
	"context"
	"database/sql"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// TestUserInfo 测试用户信息
var TestUserInfo = struct {
	ID       uint64
	Email    string
	Password string
}{
	Email:    "test@example.com",
	Password: "testpassword",
}

// GetTestConfig 获取测试配置
func GetTestConfig() (*config.Config, error) {
	// 直接设置环境变量，确保数据库连接信息被正确传递
	setTestEnv()

	// 加载配置
	return config.LoadConfig()
}

// setTestEnv 设置测试环境变量
func setTestEnv() {
	// 设置数据库配置
	if os.Getenv("DB_TYPE") == "" {
		os.Setenv("DB_TYPE", "mysql")
	}
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_USERNAME", "root")
	os.Setenv("DB_PASSWORD", "231966")
	os.Setenv("DB_DATABASE", "assistant_qisumi")
	if os.Getenv("DB_FILE_PATH") == "" {
		os.Setenv("DB_FILE_PATH", "test_assistant.db")
	}

	// 设置其他必要的环境变量
	os.Setenv("JWT_SECRET", "dev-secret-key-for-jwt-authentication")
	os.Setenv("JWT_EXPIRE_HOUR", "24")
	os.Setenv("API_KEY_ENCRYPTION_KEY", "dev-32-byte-encryption-key-for-api-keys")
}

// SetupTestUser 设置测试用户
func SetupTestUser(dbConn *sql.DB) error {
	// 检查用户是否已存在
	var count int
	err := dbConn.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM users WHERE email = ?`, TestUserInfo.Email).Scan(&count)
	if err != nil {
		return err
	}

	// 如果用户已存在，直接返回
	if count > 0 {
		// 获取用户ID
		err := dbConn.QueryRowContext(context.Background(),
			`SELECT id FROM users WHERE email = ?`, TestUserInfo.Email).Scan(&TestUserInfo.ID)
		return err
	}

	// 创建密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(TestUserInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建用户
	result, err := dbConn.ExecContext(context.Background(),
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		TestUserInfo.Email, string(hash),
	)
	if err != nil {
		return err
	}

	// 获取用户ID
	userID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	TestUserInfo.ID = uint64(userID)
	return nil
}
