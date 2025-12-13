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
	log.Println("=== Запуск Платформы Единство ===")

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
	log.Printf("Сервер запущен на http://localhost%s", config.Server.Port)
	if err := r.Run(config.Server.Port); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

func setupRouter(repo *models.Repository, secret string) *gin.Engine {
	r := gin.Default()
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*.html")

	// Используем RegisterRoutes из web.go
	handlers.RegisterRoutes(r, repo, secret)

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
