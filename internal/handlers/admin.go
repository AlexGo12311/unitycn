package handlers

import (
	"net/http"
	"strconv"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

// === ADMIN HANDLERS ===
func AdminDashboard(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := repo.GetStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения статистики"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"stats":   stats,
			"message": "Добро пожаловать в панель администратора",
		})
	}
}

func AdminPosts(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := repo.GetPosts(100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения постов"})
			return
		}

		c.JSON(http.StatusOK, posts)
	}
}

func CreatePostAdmin(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID  int    `json:"user_id"`
			Content string `json:"content"`
			Slogan  string `json:"slogan"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		if req.Slogan == "" {
			req.Slogan = "团结"
		}

		err := repo.CreatePost(req.UserID, req.Content, req.Slogan)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания поста"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Пост создан администратором"})
	}
}

func DeletePost(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		err = repo.DeletePost(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления поста"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Пост удален"})
	}
}

func AdminHeroes(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		heroes, err := repo.GetHeroes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения героев"})
			return
		}

		c.JSON(http.StatusOK, heroes)
	}
}

func GetUsers(repo *models.Repository) gin.HandlerFunc {
	// Заглушка - нужно добавить метод GetUsers в репозиторий
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"message": "В разработке"})
	}
}
