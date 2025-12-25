package http

import (
	"errors"
	"strconv"

	"assistant-qisumi/internal/auth"
	"assistant-qisumi/internal/domain"

	"github.com/gin-gonic/gin"
)

// ParseUint64Param 解析URL参数为uint64
func ParseUint64Param(c *gin.Context, paramName string) (uint64, error) {
	str := c.Param(paramName)
	id, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		R.BadRequest(c, "invalid "+paramName)
		return 0, err
	}
	return id, nil
}

// GetLLMConfig 获取LLM配置
// 现在 auth.LLMConfig 和 domain.LLMConfig 是同一类型，无需类型转换
func GetLLMConfig(c *gin.Context, svc *auth.LLMSettingService, userID uint64) (*domain.LLMConfig, error) {
	llmConfig, err := svc.GetLLMConfig(c.Request.Context(), userID)
	if err != nil {
		R.InternalError(c, "failed to get LLM config")
		return nil, err
	}
	if llmConfig == nil || llmConfig.APIKey == "" {
		R.BadRequest(c, "LLM API key not set. Please configure it in settings or contact administrator.")
		return nil, errors.New("LLM API key not set")
	}
	return llmConfig, nil
}
