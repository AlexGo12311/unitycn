package handlers

import (
	"log"
	"net/http"
	"strings"
	"unitycn/internal/auth"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(repo *models.Repository, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Проверяем заголовок Authorization (для API)
		tokenString = c.GetHeader("Authorization")

		// 2. Если нет в заголовке - проверяем куки (для веб-страниц)
		if tokenString == "" {
			tokenString, _ = c.Cookie("token")
		}

		// 3. Если вообще нет токена - ошибка
		if tokenString == "" {
			// Определяем тип запроса
			isAPI := strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") ||
				strings.Contains(c.Request.Header.Get("Accept"), "application/json") ||
				strings.HasPrefix(c.Request.URL.Path, "/api/")

			if isAPI {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			} else {
				// Для веб-страниц перенаправляем на логин
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// Удаляем префикс "Bearer " если есть
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := auth.VerifyToken(tokenString)
		if err != nil {
			// Удаляем невалидную куку
			c.SetCookie("token", "", -1, "/", "", false, true)

			isAPI := strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") ||
				strings.Contains(c.Request.Header.Get("Accept"), "application/json") ||
				strings.HasPrefix(c.Request.URL.Path, "/api/")

			if isAPI {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
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

// OptionalAuthMiddleware - НЕОБЯЗАТЕЛЬНАЯ аутентификация (не прерывает для гостей)
func OptionalAuthMiddleware(repo *models.Repository, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Проверяем заголовок Authorization (для API)
		tokenString = c.GetHeader("Authorization")

		// 2. Если нет в заголовке - проверяем куки (для веб-страниц)
		if tokenString == "" {
			tokenString, _ = c.Cookie("token")
		}

		// 3. Если нет токена - просто продолжаем как гость
		if tokenString == "" {
			c.Next()
			return
		}

		// Удаляем префикс "Bearer " если есть
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims, err := auth.VerifyToken(tokenString)
		if err != nil {
			// Невалидный токен - удаляем куку и продолжаем как гость
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Next()
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
