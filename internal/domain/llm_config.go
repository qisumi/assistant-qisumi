package domain

// LLMConfig 统一的 LLM 配置结构
// 用于存储和传递 LLM API 调用所需的配置信息
type LLMConfig struct {
	BaseURL         string `json:"base_url"`
	APIKey          string `json:"api_key,omitempty"` // 明文API密钥，仅用于客户端调用
	Model           string `json:"model"`
	ThinkingType    string `json:"thinking_type"`    // disabled, enabled, auto
	ReasoningEffort string `json:"reasoning_effort"` // low, medium, high, minimal
	EnableThinking  bool   `json:"enable_thinking"`  // true/false
	AssistantName   string `json:"assistant_name"`   // 助手名称
	HasAPIKey       bool   `json:"has_api_key"`      // 是否已设置 API Key
	IsDefault       bool   `json:"is_default"`       // 是否使用默认配置
}

// LLMSettingRequest 创建或更新 LLM 配置的请求
type LLMSettingRequest struct {
	BaseURL         string `json:"base_url" binding:"required"`
	APIKey          string `json:"api_key"` // 可选：为空时表示不修改已有API Key
	Model           string `json:"model" binding:"required"`
	ThinkingType    string `json:"thinking_type"`    // disabled, enabled, auto
	ReasoningEffort string `json:"reasoning_effort"` // low, medium, high, minimal
	EnableThinking  bool   `json:"enable_thinking"`  // true/false
	AssistantName   string `json:"assistant_name"`   // 助手名称
}

// ToClientConfig 转换为用于 LLM 客户端调用的配置（不含额外字段）
func (c *LLMConfig) ToClientConfig() *LLMConfig {
	return &LLMConfig{
		BaseURL:         c.BaseURL,
		APIKey:          c.APIKey,
		Model:           c.Model,
		ThinkingType:    c.ThinkingType,
		ReasoningEffort: c.ReasoningEffort,
		EnableThinking:  c.EnableThinking,
	}
}
