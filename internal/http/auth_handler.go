package http

import (
	"assistant-qisumi/internal/auth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", h.register)
	rg.POST("/login", h.login)
}

type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		R.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.Register(c, req.Email, req.Password); err != nil {
		R.BadRequest(c, err.Error())
		return
	}
	R.Created(c, nil)
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		R.BadRequest(c, err.Error())
		return
	}
	token, err := h.svc.Login(c, req.Email, req.Password)
	if err != nil {
		R.Unauthorized(c, err.Error())
		return
	}
	R.Success(c, gin.H{"token": token})
}
