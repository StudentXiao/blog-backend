package handler

import (
	"blog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	user, token, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	user, token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}
