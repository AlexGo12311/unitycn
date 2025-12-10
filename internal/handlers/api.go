package handlers

import (
	"net/http"
	"strconv"
	"time"
	"unitycn/internal/auth"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

// === AUTH HANDLERS ===
func Login(repo *models.Repository, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		user, err := repo.GetUserByUsername(req.Username)
		if err != nil || !auth.CheckPasswordHash(req.Password, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
			return
		}

		token, err := auth.GenerateToken(user.Username, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
			},
		})
	}
}

func Register(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хэширования пароля"})
			return
		}

		err = repo.CreateUser(req.Username, hashedPassword, "user")
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь уже существует"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Пользователь создан"})
	}
}

// === POST HANDLERS ===
func CreatePost(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		var req struct {
			Content string `json:"content"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		err := repo.CreatePost(userID, req.Content, "团结")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания поста"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Пост создан"})
	}
}

func GetPosts(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := repo.GetPosts(50, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения постов"})
			return
		}

		c.JSON(http.StatusOK, posts)
	}
}

func LikePost(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		err = repo.LikePost(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка лайка"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Лайк добавлен"})
	}
}

// === HERO HANDLERS ===
func GetHeroes(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		heroes, err := repo.GetHeroes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения героев"})
			return
		}

		c.JSON(http.StatusOK, heroes)
	}
}

func CreateHero(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			ImageURL    string `json:"image_url"`
			BirthDate   string `json:"birth_date"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты. Используйте YYYY-MM-DD"})
			return
		}

		err = repo.CreateHero(req.Name, req.Description, req.ImageURL, birthDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания героя"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Герой создан"})
	}
}

func UpdateHero(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			ImageURL    string `json:"image_url"`
			BirthDate   string `json:"birth_date"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты"})
			return
		}

		err = repo.UpdateHero(id, req.Name, req.Description, req.ImageURL, birthDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления героя"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Герой обновлен"})
	}
}

func DeleteHero(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		err = repo.DeleteHero(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления героя"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Герой удален"})
	}
}
