package auth

import (
	"context"
	"database/sql"
	"time"
)

// LLMSetting 用户LLM配置结构体
type LLMSetting struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	BaseURL   string    `json:"base_url"`
	APIKeyEnc string    `json:"api_key_enc"` // 加密后的API密钥
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LLMSettingRepository 用户LLM配置仓库
type LLMSettingRepository struct {
	db *sql.DB
}

// NewLLMSettingRepository 创建新的用户LLM配置仓库
func NewLLMSettingRepository(db *sql.DB) *LLMSettingRepository {
	return &LLMSettingRepository{db: db}
}

// GetByUserID 根据用户ID获取LLM配置
func (r *LLMSettingRepository) GetByUserID(ctx context.Context, userID uint64) (*LLMSetting, error) {
	var setting LLMSetting
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, base_url, api_key_enc, model, created_at, updated_at FROM user_llm_settings WHERE user_id = ?`,
		userID,
	).Scan(&setting.ID, &setting.UserID, &setting.BaseURL, &setting.APIKeyEnc, &setting.Model, &setting.CreatedAt, &setting.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

// Create 创建新的LLM配置
func (r *LLMSettingRepository) Create(ctx context.Context, setting *LLMSetting) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO user_llm_settings (user_id, base_url, api_key_enc, model) VALUES (?, ?, ?, ?)`,
		setting.UserID, setting.BaseURL, setting.APIKeyEnc, setting.Model,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	setting.ID = uint64(id)
	setting.CreatedAt = time.Now()
	setting.UpdatedAt = time.Now()

	return nil
}

// Update 更新LLM配置
func (r *LLMSettingRepository) Update(ctx context.Context, setting *LLMSetting) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE user_llm_settings SET base_url = ?, api_key_enc = ?, model = ? WHERE user_id = ?`,
		setting.BaseURL, setting.APIKeyEnc, setting.Model, setting.UserID,
	)

	if err != nil {
		return err
	}

	setting.UpdatedAt = time.Now()

	return nil
}

// Delete 删除LLM配置
func (r *LLMSettingRepository) Delete(ctx context.Context, userID uint64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_llm_settings WHERE user_id = ?`,
		userID,
	)

	return err
}
