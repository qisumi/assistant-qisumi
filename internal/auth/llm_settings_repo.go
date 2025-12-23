package auth

import (
	"context"

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
	var settings []UserLLMSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(1).Find(&settings).Error

	if err != nil {
		return nil, err
	}

	if len(settings) == 0 {
		return nil, nil
	}

	return &settings[0], nil
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
