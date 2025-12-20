package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// LLMSettingRepository 用户LLM配置仓库
type LLMSettingRepository struct {
	db *gorm.DB
}

// NewLLMSettingRepository 创建新的用户LLM配置仓库
func NewLLMSettingRepository(db *gorm.DB) *LLMSettingRepository {
	return &LLMSettingRepository{db: db}
}

// GetByUserID 根据用户ID获取LLM配置
func (r *LLMSettingRepository) GetByUserID(ctx context.Context, userID uint64) (*UserLLMSetting, error) {
	var setting UserLLMSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

// Create 创建新的LLM配置
func (r *LLMSettingRepository) Create(ctx context.Context, setting *UserLLMSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

// Update 更新LLM配置
func (r *LLMSettingRepository) Update(ctx context.Context, setting *UserLLMSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

// Delete 删除LLM配置
func (r *LLMSettingRepository) Delete(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserLLMSetting{}).Error
}
