package handlers

import "github.com/gin-gonic/gin"

// APIError — единая форма ошибки, которую возвращают handler'ы.
// Code — машинно-читаемый идентификатор (snake_case), Message — человекочитаемая
// строка для UI. Field, при необходимости, указывает на конкретное поле формы.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// abortWithError — короткая запись для гин-обработчиков. Гарантированно
// останавливает цепочку (AbortWithStatusJSON) и отдаёт фронту устойчивый shape
// `{ "error": { code, message, field? } }`.
//
// Совместимость со старым клиентом: дублируем `error: <message>` верхнего уровня,
// чтобы FE, который ещё читает err.message ?? err.error, продолжал работать.
func abortWithError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error":   message,
		"code":    code,
		"message": message,
		"details": APIError{Code: code, Message: message},
	})
}

// Часто используемые коды. Не enum, чтобы можно было свободно добавлять.
const (
	ErrCodeInvalidJSON        = "invalid_json"
	ErrCodeInvalidUUID        = "invalid_uuid"
	ErrCodeInvalidUserID      = "invalid_user_id"
	ErrCodeUnauthorized       = "unauthorized"
	ErrCodeQuizNotFound       = "quiz_not_found"
	ErrCodeReportNotFound     = "report_not_found"
	ErrCodeSessionNotFound    = "session_not_found"
	ErrCodeInvalidOldPassword = "invalid_old_password"
	ErrCodeInternal           = "internal"
)
