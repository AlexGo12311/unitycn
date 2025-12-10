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
	// Включим подробное логирование
	log.Println("=== Запуск Communist Twitter ===")

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

	log.Printf("Конфиг загружен: порт=%s", config.Server.Port)

	// Подключение к БД
	log.Println("Подключение к PostgreSQL...")
	db, err := database.Connect(config.Database)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Инициализация репозитория
	repo := models.NewRepository(db)

	// Инициализация аутентификации
	auth.Init(config.Server.Secret)
	log.Println("Аутентификация инициализирована")

	// Создание админа, если его нет
	log.Printf("Проверка администратора: %s", config.Admin.Username)
	if err := ensureAdminExists(repo, config.Admin.Username, config.Admin.Password); err != nil {
		log.Printf("Предупреждение: не удалось создать администратора: %v", err)
	} else {
		log.Println("Администратор проверен/создан")
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
	r.GET("/register", handlers.RegisterPage())

	// API
	api := r.Group("/api")
	{
		api.POST("/login", handlers.Login(repo, secret))
		api.POST("/register", handlers.Register(repo))
		api.GET("/posts", handlers.GetPosts(repo))
		api.GET("/heroes", handlers.GetHeroes(repo))

		// Требуется авторизация - ПЕРЕДАЁМ РЕПОЗИТОРИЙ
		authApi := api.Group("")
		authApi.Use(handlers.AuthMiddleware(repo, secret)) // <- ВАЖНО: передаём repo
		{
			authApi.POST("/posts", handlers.CreatePost(repo))
			authApi.POST("/posts/:id/like", handlers.LikePost(repo))
		}
	}

	// Админка - ТАКЖЕ ПЕРЕДАЁМ РЕПОЗИТОРИЙ
	admin := r.Group("/admin")
	admin.Use(handlers.AuthMiddleware(repo, secret), handlers.AdminMiddleware()) // <- передаём repo
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
		log.Printf("Администратор %s уже существует", username)
		return nil
	}

	log.Printf("Создание администратора: %s", username)

	// Если пользователя нет - создаём с display_name = username
	return repo.CreateUser(username, hashedPassword, "admin", username)
}
