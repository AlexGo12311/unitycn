package handlers

import (
	"net/http"
	"unitycn/internal/models"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, repo *models.Repository, secret string) {
	// Веб-страницы
	r.GET("/", HomePage(repo))
	r.GET("/login", LoginPage())

	// API endpoints
	api := r.Group("/api")
	{
		api.POST("/login", Login(repo, secret))
		api.POST("/register", Register(repo))
		api.GET("/posts", GetPosts(repo))
		api.GET("/heroes", GetHeroes(repo))

		// Требуется авторизация - передаём репозиторий
		authApi := api.Group("")
		authApi.Use(AuthMiddleware(repo, secret)) // <- добавляем repo
		{
			authApi.POST("/posts", CreatePost(repo))
			authApi.POST("/posts/:id/like", LikePost(repo))

			authApi.POST("/posts/:post_id/comments", CreateComment(repo))
			authApi.GET("/posts/:post_id/comments", GetComments(repo))
			authApi.DELETE("/comments/:id", DeleteComment(repo))
		}
	}

	// Админка
	admin := r.Group("/admin")
	admin.Use(AuthMiddleware(repo, secret), AdminMiddleware())
	{
		admin.GET("/", AdminDashboard(repo))
		admin.GET("/posts", AdminPosts(repo))
		admin.POST("/posts", CreatePostAdmin(repo))
		admin.DELETE("/posts/:id", DeletePost(repo))

		admin.GET("/heroes", AdminHeroes(repo))
		admin.POST("/heroes", CreateHero(repo))
		admin.PUT("/heroes/:id", UpdateHero(repo))
		admin.DELETE("/heroes/:id", DeleteHero(repo))

		admin.GET("/users", GetUsers(repo))
	}
}

func HomePage(repo *models.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {

		posts, _ := repo.GetPostsWithUsers(10, 0)
		heroes, _ := repo.GetHeroes()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":  "Единство 团结 - Пролетарская платформа",
			"slogan": "Пролетарии всех стран, соединяйтесь!",
			"posts":  posts,
			"heroes": heroes,
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
