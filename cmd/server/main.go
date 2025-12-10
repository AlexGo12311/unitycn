package main

import (
	"log"

	"unitycn/internal/auth"
	"unitycn/internal/database"
	"unitycn/internal/handlers"
	"unitycn/internal/models"

	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port   string `yaml:"port"`
		Secret string `yaml:"secret_key"`
	} `yaml:"server"`
	Database database.DBConfig `yaml:"database"`
	Admin    struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"admin"`
}

func main() {
	// Чтение конфига
	configFile, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal("Ошибка открытия config.yaml:", err)
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("Ошибка парсинга config.yaml:", err)
	}

	// Подключение к БД
	db, err := database.Connect(config.Database)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Инициализация репозитория
	repo := models.NewRepository(db)

	// Инициализация аутентификации
	auth.Init(config.Server.Secret)

	// Создание админа, если его нет
	if err := ensureAdminExists(repo, config.Admin.Username, config.Admin.Password); err != nil {
		log.Printf("Предупреждение: не удалось создать администратора: %v", err)
	}

	// Настройка маршрутов
	r := setupRouter(repo, config.Server.Secret)

	// Запуск сервера
	log.Printf("🚀 Сервер запущен на http://localhost%s", config.Server.Port)
	if err := r.Run(config.Server.Port); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

func setupRouter(repo *models.Repository, secret string) *gin.Engine {
	r := gin.Default()

	// Статические файлы
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*.html")

	// Основные маршруты
	r.GET("/", handlers.HomePage(repo))
	r.GET("/login", handlers.LoginPage())

	// API
	api := r.Group("/api")
	{
		api.POST("/login", handlers.Login(repo, secret))
		api.POST("/register", handlers.Register(repo))
		api.GET("/posts", handlers.GetPosts(repo))
		api.GET("/heroes", handlers.GetHeroes(repo))

		// Требуется авторизация
		authApi := api.Group("")
		authApi.Use(handlers.AuthMiddleware(secret))
		{
			authApi.POST("/posts", handlers.CreatePost(repo))
			authApi.POST("/posts/:id/like", handlers.LikePost(repo))
		}
	}

	// Админка
	admin := r.Group("/admin")
	admin.Use(handlers.AuthMiddleware(secret), handlers.AdminMiddleware())
	{
		admin.GET("/", handlers.AdminDashboard(repo))
		admin.GET("/posts", handlers.AdminPosts(repo))
		admin.POST("/posts", handlers.CreatePostAdmin(repo))
		admin.DELETE("/posts/:id", handlers.DeletePost(repo))

		admin.GET("/heroes", handlers.AdminHeroes(repo))
		admin.POST("/heroes", handlers.CreateHero(repo))
		admin.PUT("/heroes/:id", handlers.UpdateHero(repo))
		admin.DELETE("/heroes/:id", handlers.DeleteHero(repo))

		admin.GET("/users", handlers.GetUsers(repo))
	}

	return r
}

func ensureAdminExists(repo *models.Repository, username, hashedPassword string) error {
	// Проверяем, существует ли уже администратор
	_, err := repo.GetUserByUsername(username)
	if err == nil {
		// Пользователь уже существует
		return nil
	}

	// Если пользователя нет - создаём
	return repo.CreateUser(username, hashedPassword, "admin")
}
