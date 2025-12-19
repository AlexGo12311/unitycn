package handlers

import (
	"net/http"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, repo *models.Repository, secret string) {
	// Применяем OptionalAuthMiddleware глобально ко всем маршрутам
	r.Use(OptionalAuthMiddleware(repo, secret))

	// Веб-страницы
	r.GET("/", HomePage(repo))
	r.GET("/login", LoginPage())
	r.GET("/register", RegisterPage())
	r.GET("/logout", Logout())

	// ВЕБ-форма логина
	r.POST("/login", Login(repo, secret))

	// ВЕБ-форма регистрации
	r.POST("/register", Register(repo, secret))

	// API endpoints
	api := r.Group("/api")
	{
		api.POST("/login", Login(repo, secret))
		api.POST("/register", Register(repo, secret))
		api.GET("/posts", GetPosts(repo))
		api.GET("/heroes", GetHeroes(repo))

		// Требуется авторизация (используем строгий AuthMiddleware)
		authApi := api.Group("")
		authApi.Use(AuthMiddleware(repo, secret))
		{
			authApi.POST("/posts", CreatePost(repo))
			authApi.POST("/posts/:id/like", LikePost(repo))
			authApi.POST("/posts/:id/comments", CreateComment(repo))
			authApi.GET("/posts/:id/comments", GetComments(repo))
			authApi.DELETE("/comments/:id", DeleteComment(repo))
		}
	}

	// Админка требует строгой авторизации
	admin := r.Group("/admin")
	admin.Use(AuthMiddleware(repo, secret), AdminMiddleware())
	{
		// Дашборд
		admin.GET("/", AdminDashboard(repo))

		// Пользователи
		admin.GET("/users", AdminUsers(repo))
		admin.PUT("/users/:id/role", UpdateUserRole(repo))
		admin.DELETE("/users/:id", DeleteUserAdmin(repo))

		// Посты
		admin.GET("/posts", AdminPosts(repo))
		admin.GET("/posts/:id/edit", EditPostPage(repo))
		admin.PUT("/posts/:id", UpdatePostAdmin(repo))
		admin.DELETE("/posts/:id", DeletePostAdmin(repo))

		// Комментарии
		admin.GET("/comments", AdminComments(repo))
		admin.DELETE("/comments/:id", DeleteCommentAdmin(repo))

		// Герои
		admin.GET("/heroes", AdminHeroes(repo))
		admin.POST("/heroes", CreateHero(repo))
		admin.PUT("/heroes/:id", UpdateHero(repo))
		admin.DELETE("/heroes/:id", DeleteHero(repo))
		admin.GET("/heroes/:id", GetHeroByID(repo))
	}
}

// Другие функции идут ПОСЛЕ закрывающей фигурной скобки
func HomePage(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем пользователя из контекста (AuthMiddleware должен был его установить)
		var userObj *models.User

		// Способ 1: Проверяем, установлен ли user в контексте
		if userData, exists := c.Get("user"); exists {
			if user, ok := userData.(*models.User); ok {
				userObj = user
			}
		}

		// Способ 2: Если нет в контексте, пытаемся получить по user_id
		if userObj == nil {
			if userID, exists := c.Get("user_id"); exists {
				if userIDInt, ok := userID.(int); ok && userIDInt > 0 {
					// Получаем пользователя из БД
					user, err := repo.GetUserByID(userIDInt)
					if err == nil {
						userObj = user
					}
				}
			}
		}

		posts, _ := repo.GetPostsWithUsers(10, 0)
		heroes, _ := repo.GetHeroes()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":  "Единство 团结 - Пролетарская платформа",
			"slogan": "Пролетарии всех стран, соединяйтесь!",
			"posts":  posts,
			"heroes": heroes,
			"user":   userObj,
		})
	}
}

func LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	}
}

func RegisterPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	}
}
