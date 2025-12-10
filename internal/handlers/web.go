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

		// Требуется авторизация
		authApi := api.Group("")
		authApi.Use(AuthMiddleware(secret))
		{
			authApi.POST("/posts", CreatePost(repo))
			authApi.POST("/posts/:id/like", LikePost(repo))
		}
	}

	// Админка
	admin := r.Group("/admin")
	admin.Use(AuthMiddleware(secret), AdminMiddleware())
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
		posts, _ := repo.GetPosts(10, 0)
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
