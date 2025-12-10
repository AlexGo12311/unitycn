package models

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// === USERS ===
func (r *Repository) CreateUser(username, password, role, displayName string) error {
	if displayName == "" {
		displayName = username // по умолчанию
	}

	query := `INSERT INTO users (username, password, role, display_name) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, username, password, role, displayName)
	return err
}

func (r *Repository) GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, display_name, password, role, created_at FROM users WHERE username = $1`
	row := r.db.QueryRow(query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// === POSTS ===
func (r *Repository) CreatePost(userID int, content, slogan string) error {
	query := `INSERT INTO posts (user_id, content, slogan) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, userID, content, slogan)
	return err
}

func (r *Repository) GetPosts(limit, offset int) ([]Post, error) {
	query := `
        SELECT p.id, p.user_id, p.content, p.slogan, p.likes, p.created_at,
               u.id, u.username, u.role, u.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var user User
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.Slogan, &post.Likes, &post.CreatedAt,
			&user.ID, &user.Username, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		post.User = &user
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *Repository) UpdatePost(id int, content string) error {
	query := `UPDATE posts SET content = $1 WHERE id = $2`
	_, err := r.db.Exec(query, content, id)
	return err
}

func (r *Repository) DeletePost(id int) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *Repository) LikePost(postID, userID int) (bool, error) {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Проверяем, ставил ли пользователь уже лайк
	var exists bool
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)",
		postID, userID,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	if exists {
		// Убираем лайк (дизлайк)
		_, err = tx.Exec("DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2", postID, userID)
		if err != nil {
			return false, err
		}

		// Уменьшаем счётчик лайков
		_, err = tx.Exec("UPDATE posts SET likes = likes - 1 WHERE id = $1", postID)
		if err != nil {
			return false, err
		}

		err = tx.Commit()
		return false, err // false = лайк убран

	} else {
		// Добавляем лайк
		_, err = tx.Exec(
			"INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2)",
			postID, userID,
		)
		if err != nil {
			return false, err
		}

		// Увеличиваем счётчик лайков
		_, err = tx.Exec("UPDATE posts SET likes = likes + 1 WHERE id = $1", postID)
		if err != nil {
			return false, err
		}

		err = tx.Commit()
		return true, err // true = лайк добавлен
	}
}

// Получить статус лайка пользователя
func (r *Repository) GetUserLikeStatus(postID, userID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)",
		postID, userID,
	).Scan(&exists)

	return exists, err
}

// Получить количество лайков поста
func (r *Repository) GetPostLikesCount(postID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT likes FROM posts WHERE id = $1", postID).Scan(&count)
	return count, err
}

// Получение по пользователю
func (r *Repository) GetPostsWithUsers(limit, offset int) ([]Post, error) {
	query := `
        SELECT p.id, p.user_id, p.content, p.slogan, p.likes, p.created_at,
               u.id, u.username, u.display_name, u.role, u.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var user User
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.Slogan, &post.Likes, &post.CreatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		post.User = &user
		posts = append(posts, post)
	}
	return posts, nil
}

// === HEROES (второй объект) ===
func (r *Repository) CreateHero(name, description, imageURL string, birthDate time.Time) error {
	query := `INSERT INTO heroes (name, description, birth_date, image_url) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, name, description, birthDate, imageURL)
	return err
}

func (r *Repository) GetHeroes() ([]Hero, error) {
	query := `SELECT id, name, description, birth_date, image_url, created_at FROM heroes ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var heroes []Hero
	for rows.Next() {
		var hero Hero
		err := rows.Scan(&hero.ID, &hero.Name, &hero.Description, &hero.BirthDate, &hero.ImageURL, &hero.CreatedAt)
		if err != nil {
			return nil, err
		}
		heroes = append(heroes, hero)
	}
	return heroes, nil
}

func (r *Repository) UpdateHero(id int, name, description, imageURL string, birthDate time.Time) error {
	query := `UPDATE heroes SET name=$1, description=$2, birth_date=$3, image_url=$4 WHERE id=$5`
	_, err := r.db.Exec(query, name, description, birthDate, imageURL, id)
	return err
}

func (r *Repository) DeleteHero(id int) error {
	query := `DELETE FROM heroes WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// === ADMIN STATISTICS ===
func (r *Repository) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	// Количество пользователей
	var userCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return nil, err
	}
	stats["users"] = userCount

	// Количество постов
	var postCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	if err != nil {
		return nil, err
	}
	stats["posts"] = postCount

	// Количество героев
	var heroCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM heroes").Scan(&heroCount)
	if err != nil {
		return nil, err
	}
	stats["heroes"] = heroCount

	return stats, nil
}
