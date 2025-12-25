package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// LLMSettingService 用户LLM配置服务
type LLMSettingService struct {
	repo          *LLMSettingRepository
	encryptionKey []byte
	defaultConfig *LLMConfig
}

// LLMConfig 用于对外暴露的LLM配置，不包含加密的API密钥
type LLMConfig struct {
	BaseURL          string `json:"base_url"`
	APIKey           string `json:"api_key,omitempty"` // 明文API密钥，仅用于客户端调用
	Model            string `json:"model"`
	ThinkingType     string `json:"thinking_type"`      // disabled, enabled, auto
	ReasoningEffort  string `json:"reasoning_effort"`   // low, medium, high, minimal
	AssistantName    string `json:"assistant_name"`     // 助手名称
	HasAPIKey        bool   `json:"has_api_key"`
	IsDefault        bool   `json:"is_default"`
}

// LLMSettingRequest 创建或更新LLM配置的请求
type LLMSettingRequest struct {
	BaseURL         string `json:"base_url" binding:"required"`
	APIKey          string `json:"api_key"`           // 可选：为空时表示不修改已有API Key
	Model           string `json:"model" binding:"required"`
	ThinkingType    string `json:"thinking_type"`     // disabled, enabled, auto
	ReasoningEffort string `json:"reasoning_effort"`  // low, medium, high, minimal
	AssistantName   string `json:"assistant_name"`    // 助手名称
}

// NewLLMSettingService 创建新的用户LLM配置服务
func NewLLMSettingService(repo *LLMSettingRepository, encryptionKey string, defaultConfig *LLMConfig) *LLMSettingService {
	return &LLMSettingService{
		repo:          repo,
		encryptionKey: []byte(encryptionKey),
		defaultConfig: defaultConfig,
	}
}

// encryptAPIKey 加密API密钥
func (s *LLMSettingService) encryptAPIKey(apiKey string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptAPIKey 解密API密钥
func (s *LLMSettingService) decryptAPIKey(encryptedKey string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("invalid ciphertext")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetLLMConfig 获取用户的LLM配置
func (s *LLMSettingService) GetLLMConfig(ctx context.Context, userID uint64) (*LLMConfig, error) {
	setting, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if setting == nil {
		// 如果用户没有配置，返回默认配置
		return s.defaultConfig, nil
	}

	apiKey, err := s.decryptAPIKey(setting.APIKeyEnc)
	if err != nil {
		return nil, err
	}

	return &LLMConfig{
		BaseURL:         setting.BaseURL,
		APIKey:          apiKey,
		Model:           setting.Model,
		ThinkingType:    setting.ThinkingType,
		ReasoningEffort: setting.ReasoningEffort,
		AssistantName:   setting.AssistantName,
		HasAPIKey:       true,
		IsDefault:       false,
	}, nil
}

// UpdateLLMSetting 更新用户的LLM配置
func (s *LLMSettingService) UpdateLLMSetting(ctx context.Context, userID uint64, req LLMSettingRequest) error {
	// 检查是否已存在配置
	existingSetting, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	var encryptedKey string

	if existingSetting == nil {
		// 创建新配置：API Key 必须提供
		if req.APIKey == "" {
			return errors.New("API Key is required when creating new settings")
		}
		// 加密API密钥
		encryptedKey, err = s.encryptAPIKey(req.APIKey)
		if err != nil {
			return err
		}

		setting := &UserLLMSetting{
			UserID:          userID,
			BaseURL:         req.BaseURL,
			APIKeyEnc:       encryptedKey,
			Model:           req.Model,
			ThinkingType:    req.ThinkingType,
			ReasoningEffort: req.ReasoningEffort,
			AssistantName:   req.AssistantName,
		}
		return s.repo.Create(ctx, setting)
	}

	// 更新现有配置
	existingSetting.BaseURL = req.BaseURL
	existingSetting.Model = req.Model
	existingSetting.ThinkingType = req.ThinkingType
	existingSetting.ReasoningEffort = req.ReasoningEffort
	existingSetting.AssistantName = req.AssistantName

	// 仅在提供了新 API Key 时才更新
	if req.APIKey != "" {
		encryptedKey, err = s.encryptAPIKey(req.APIKey)
		if err != nil {
			return err
		}
		existingSetting.APIKeyEnc = encryptedKey
	}

	return s.repo.Update(ctx, existingSetting)
}

// DeleteLLMSetting 删除用户的LLM配置
func (s *LLMSettingService) DeleteLLMSetting(ctx context.Context, userID uint64) error {
	return s.repo.Delete(ctx, userID)
}
