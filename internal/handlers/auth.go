package handlers

import (
	"log"
	"net/http"
	"unitycn/internal/auth"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(repo *models.Repository, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := auth.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		username := claims["username"].(string)
		role := claims["role"].(string)

		user, err := repo.GetUserByUsername(username)
		if err != nil {
			log.Printf("Предупреждение: не удалось получить пользователя %s: %v", username, err)
			c.Set("username", username)
			c.Set("role", role)
			c.Set("user_id", 0)
		} else {
			c.Set("username", username)
			c.Set("role", role)
			c.Set("user_id", user.ID)
			c.Set("user", user)
		}

		c.Next()
	}
}

// Альтернативная версия без репозитория (если хранить user_id в токене)
func AuthMiddlewareSimple(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := auth.VerifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("username", claims["username"].(string))
		c.Set("role", claims["role"].(string))

		// Пытаемся получить user_id из токена (если он там есть)
		if userID, ok := claims["user_id"].(float64); ok {
			c.Set("user_id", int(userID))
		} else {
			c.Set("user_id", 0)
		}

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Требуются права администратора"})
			c.Abort()
			return
		}
		c.Next()
	}
}
