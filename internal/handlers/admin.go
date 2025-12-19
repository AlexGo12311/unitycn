package handlers

import (
	"net/http"
	"strconv"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

// === АДМИН HANDLERS ===

// AdminDashboard - главная страница админки
func AdminDashboard(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := repo.GetStats()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
				"error": "Ошибка получения статистики",
			})
			return
		}

		c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
			"title": "Админ-панель",
			"stats": stats,
		})
	}
}

// AdminUsers - управление пользователями
func AdminUsers(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repo.GetAllUsers()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
				"error": "Ошибка получения пользователей",
			})
			return
		}

		c.HTML(http.StatusOK, "admin/users.html", gin.H{
			"title": "Управление пользователями",
			"users": users,
		})
	}
}

// UpdateUserRole - обновление роли пользователя
func UpdateUserRole(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		var req struct {
			Role string `json:"role"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		// Проверяем, что роль валидная
		if req.Role != "admin" && req.Role != "user" && req.Role != "moderator" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверная роль"})
			return
		}

		err = repo.UpdateUserRole(userID, req.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления роли"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Роль обновлена",
			"role":    req.Role,
		})
	}
}

// DeleteUserAdmin - удаление пользователя (админ)
func DeleteUserAdmin(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		// Не позволяем удалить самого себя
		currentUserID, _ := c.Get("user_id")
		if currentUserID.(int) == userID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя удалить самого себя"})
			return
		}

		err = repo.DeleteUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления пользователя"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Пользователь удален",
		})
	}
}

// AdminPosts - управление постами
func AdminPosts(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, err := repo.GetAllPostsAdmin()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
				"error": "Ошибка получения постов",
			})
			return
		}

		c.HTML(http.StatusOK, "admin/posts.html", gin.H{
			"title": "Управление постами",
			"posts": posts,
		})
	}
}

// DeletePostAdmin - удаление поста (админ)
func DeletePostAdmin(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		postID, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		err = repo.DeletePostAdmin(postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления поста"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Пост удален",
		})
	}
}

// EditPostPage - страница редактирования поста
func EditPostPage(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		postID, err := strconv.Atoi(idStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "admin/error.html", gin.H{
				"error": "Неверный ID поста",
			})
			return
		}

		post, err := repo.GetPostByID(postID)
		if err != nil {
			c.HTML(http.StatusNotFound, "admin/error.html", gin.H{
				"error": "Пост не найден",
			})
			return
		}

		c.HTML(http.StatusOK, "admin/edit_post.html", gin.H{
			"title": "Редактирование поста",
			"post":  post,
		})
	}
}

// UpdatePostAdmin - обновление поста (админ)
func UpdatePostAdmin(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		postID, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		var req struct {
			Content string `json:"content"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		if req.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Содержание не может быть пустым"})
			return
		}

		err = repo.UpdatePost(postID, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления поста"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Пост обновлен",
		})
	}
}

// AdminComments - управление комментариями
func AdminComments(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		comments, err := repo.GetAllCommentsAdmin()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
				"error": "Ошибка получения комментариев",
			})
			return
		}

		c.HTML(http.StatusOK, "admin/comments.html", gin.H{
			"title":    "Управление комментариями",
			"comments": comments,
		})
	}
}

// DeleteCommentAdmin - удаление комментария (админ)
func DeleteCommentAdmin(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		commentID, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		err = repo.DeleteCommentAdmin(commentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления комментария"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Комментарий удален",
		})
	}
}

// AdminHeroes - управление героями (уже есть, обновим)
func AdminHeroes(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		heroes, err := repo.GetHeroes()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
				"error": "Ошибка получения героев",
			})
			return
		}

		c.HTML(http.StatusOK, "admin/heroes.html", gin.H{
			"title":  "Управление героями",
			"heroes": heroes,
		})
	}
}

// GetHeroByID - получение героя по ID
// GetHeroByID - получение героя по ID
func GetHeroByID(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		_, err := strconv.Atoi(idStr) // Используем переменную
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		// TODO: Добавить метод GetHeroByID в репозиторий
		c.JSON(http.StatusOK, gin.H{
			"message": "Функция в разработке",
		})
	}
}
