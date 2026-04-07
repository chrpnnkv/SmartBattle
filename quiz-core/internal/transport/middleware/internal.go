package middleware

import (
	"net/http"

	"github.com/chrpnnkv/SmartBattle/internal/config"
	"github.com/gin-gonic/gin"
)

func InternalGuard(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Internal-Secret")
		if secret != cfg.XInternalSecret {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: invalid internal secret"})
			return
		}
		c.Next()
	}
}
