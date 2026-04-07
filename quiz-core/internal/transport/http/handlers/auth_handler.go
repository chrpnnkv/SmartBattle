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
	Email    string `json:"email" binding:"required,email" example:"teacher@hse.ru"`
	Password string `json:"password" binding:"required,min=6" example:"secret123"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email" example:"teacher@hse.ru"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required" example:"secret123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newsecret123"`
}

type ForgotPasswordReq struct {
	Email string `json:"email" binding:"required,email" example:"teacher@hse.ru"`
}

type ResetPasswordReq struct {
	Token       string `json:"token" binding:"required" example:"uuid-token-from-log"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newsecret123"`
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

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
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

	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// @Summary Профиль пользователя
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")
	c.JSON(http.StatusOK, gin.H{"id": userID, "role": role})
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
