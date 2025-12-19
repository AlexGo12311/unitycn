package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"unitycn/internal/auth"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

// === AUTH HANDLERS ===

// Logout - выход из системы
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Удаляем куку с токеном
		c.SetCookie("token", "", -1, "/", "", false, true)

		// Определяем тип запроса
		isAPI := strings.Contains(c.Request.Header.Get("Accept"), "application/json") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/")

		if isAPI {
			c.JSON(http.StatusOK, gin.H{
				"message": "Вы успешно вышли из системы",
			})
		} else {
			// Для веб-страниц перенаправляем на главную
			c.Redirect(http.StatusFound, "/")
		}
	}
}

func Login(repo *models.Repository, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Определяем тип запроса
		contentType := c.ContentType()
		var username, password string

		if contentType == "application/x-www-form-urlencoded" {
			// Веб-форма
			username = c.PostForm("username")
			password = c.PostForm("password")
		} else if contentType == "application/json" {
			// API запрос
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
				return
			}

			username = req.Username
			password = req.Password
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный Content-Type"})
			return
		}

		// Проверка пользователя
		user, err := repo.GetUserByUsername(username)
		if err != nil || !auth.CheckPasswordHash(password, user.Password) {
			if contentType == "application/x-www-form-urlencoded" {
				c.HTML(http.StatusUnauthorized, "login.html", gin.H{
					"error": "Неверные учетные данные",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
			}
			return
		}

		// Генерация токена
		token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
		if err != nil {
			if contentType == "application/x-www-form-urlencoded" {
				c.HTML(http.StatusInternalServerError, "login.html", gin.H{
					"error": "Ошибка авторизации",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
			}
			return
		}

		// Устанавливаем куку в любом случае
		c.SetCookie("token", token, 24*3600, "/", "", false, true)

		// Ответ в зависимости от типа запроса
		if contentType == "application/x-www-form-urlencoded" {
			// Для веб-формы: отправляем на страницу редиректа
			c.HTML(http.StatusOK, "auth_redirect.html", gin.H{
				"token":    token,
				"username": user.Username,
				"role":     user.Role,
				"user_id":  user.ID,
				"redirect": "/",
			})
		} else {
			// Для API: возвращаем JSON
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
}

func Register(repo *models.Repository, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username    string `json:"username"`
			Password    string `json:"password"`
			DisplayName string `json:"display_name,omitempty"`
		}

		// Определяем тип запроса
		contentType := c.ContentType()

		if contentType == "application/x-www-form-urlencoded" {
			// Веб-форма
			req.Username = c.PostForm("username")
			req.Password = c.PostForm("password")
			req.DisplayName = c.PostForm("display_name")
		} else if contentType == "application/json" {
			// API запрос
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный Content-Type"})
			return
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хэширования пароля"})
			return
		}

		err = repo.CreateUser(req.Username, hashedPassword, "user", req.DisplayName)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь уже существует"})
			return
		}

		// АВТОМАТИЧЕСКАЯ АВТОРИЗАЦИЯ ПОСЛЕ РЕГИСТРАЦИИ
		// Получаем созданного пользователя
		user, err := repo.GetUserByUsername(req.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка авторизации после регистрации"})
			return
		}

		// Генерация токена
		token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
			return
		}

		// Устанавливаем куку
		c.SetCookie("token", token, 24*3600, "/", "", false, true)

		// Ответ в зависимости от типа запроса
		if contentType == "application/x-www-form-urlencoded" {
			// Для веб-формы: отправляем на страницу редиректа
			c.HTML(http.StatusOK, "auth_redirect.html", gin.H{
				"token":    token,
				"username": user.Username,
				"role":     user.Role,
				"user_id":  user.ID,
				"redirect": "/",
				"message":  "Регистрация успешна! Вы авторизованы.",
			})
		} else {
			// Для API: возвращаем JSON
			c.JSON(http.StatusCreated, gin.H{
				"message": "Пользователь создан",
				"token":   token,
				"user": gin.H{
					"id":       user.ID,
					"username": user.Username,
					"role":     user.Role,
				},
			})
		}
	}
}

// === POST HANDLERS ===
func CreatePost(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
			return
		}

		user, err := repo.GetUserByUsername(username.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
			return
		}

		var req struct {
			Content string `json:"content"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		err = repo.CreatePost(user.ID, req.Content, "团结")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания поста"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Пост создан"})
	}
}

func GetPosts(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Используем новый метод с отображением имён
		posts, err := repo.GetPostsWithUsers(50, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения постов"})
			return
		}

		c.JSON(http.StatusOK, posts)
	}
}

func LikePost(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
			return
		}

		user, err := repo.GetUserByUsername(username.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
			return
		}

		idStr := c.Param("id")
		postID, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}

		liked, err := repo.LikePost(postID, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка лайка"})
			return
		}

		message := "Лайк добавлен"
		if !liked {
			message = "Лайк убран"
		}

		c.JSON(http.StatusOK, gin.H{
			"message": message,
			"liked":   liked,
		})
	}
}

// === COMMENT HANDLERS ===

func CreateComment(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
			return
		}

		user, err := repo.GetUserByUsername(username.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
			return
		}

		// Получаем ID поста из URL параметра
		postIDStr := c.Param("id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID поста"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Комментарий не может быть пустым"})
			return
		}

		err = repo.CreateComment(postID, user.ID, req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания комментария"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Комментарий добавлен",
			"post_id": postID,
		})
	}
}

func GetComments(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		postIDStr := c.Param("id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID поста"})
			return
		}

		comments, err := repo.GetCommentsByPostID(postID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения комментариев"})
			return
		}

		c.JSON(http.StatusOK, comments)
	}
}

func DeleteComment(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
			return
		}

		user, err := repo.GetUserByUsername(username.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователя"})
			return
		}

		commentIDStr := c.Param("id")
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID комментария"})
			return
		}

		err = repo.DeleteComment(commentID, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Комментарий удален"})
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
