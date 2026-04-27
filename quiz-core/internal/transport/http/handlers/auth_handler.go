package handlers

import (
	"net/http"

	"github.com/chrpnnkv/SmartBattle/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type RegisterReq struct {
	Name     string `json:"name" binding:"required" example:"Тимофей"`
	Email    string `json:"email" binding:"required,email" example:"teacher@hse.ru"`
	Password string `json:"password" binding:"required,min=6" example:"secret123"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email" example:"teacher@hse.ru"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"oldPassword" binding:"required" example:"secret123"`
	NewPassword string `json:"newPassword" binding:"required,min=6" example:"newsecret123"`
}

type ForgotPasswordReq struct {
	Email string `json:"email" binding:"required,email" example:"teacher@hse.ru"`
}

type ResetPasswordReq struct {
	Token       string `json:"token" binding:"required" example:"uuid-token-from-log"`
	NewPassword string `json:"newPassword" binding:"required,min=6" example:"newsecret123"`
}

// @Summary Регистрация нового пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterReq true "Данные регистрации"
// @Success 201 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, token, _ := h.service.Login(req.Email, req.Password)

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		"tokens": gin.H{
			"accessToken": token,
		},
	})
}

// @Summary Вход в систему
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginReq true "Учетные данные"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		"tokens": gin.H{
			"accessToken": token,
		},
	})
}

// @Summary Профиль пользователя
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

// @Summary Смена пароля
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordReq true "Данные для смены пароля"
// @Success 200 {object} map[string]interface{}
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	if err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// @Summary Запрос на сброс пароля
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordReq true "Email пользователя"
// @Success 200 {object} map[string]interface{}
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ForgotPassword(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reset token generated (see server logs)"})
}

// @Summary Установка нового пароля по токену
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordReq true "Токен и новый пароль"
// @Success 200 {object} map[string]interface{}
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}
