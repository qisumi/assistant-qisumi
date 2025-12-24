package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一的响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseHelper 响应辅助函数
type ResponseHelper struct{}

// Success 成功响应
func (ResponseHelper) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// SuccessWithMessage 带消息的成功响应
func (ResponseHelper) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"data":    data,
	})
}

// Error 错误响应
func (ResponseHelper) Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// BadRequest 400错误
func (ResponseHelper) BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// Unauthorized 401错误
func (ResponseHelper) Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": message})
}

// Forbidden 403错误
func (ResponseHelper) Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{"error": message})
}

// NotFound 404错误
func (ResponseHelper) NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

// InternalError 500错误
func (ResponseHelper) InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

// Created 201响应
func (ResponseHelper) Created(c *gin.Context, data interface{}) {
	if data != nil {
		c.JSON(http.StatusCreated, data)
	} else {
		c.Status(http.StatusCreated)
	}
}

var R = ResponseHelper{}
