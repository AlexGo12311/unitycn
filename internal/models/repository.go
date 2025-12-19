package models

import (
	"database/sql"
	"fmt"
	"log"
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

func (r *Repository) GetUserByID(id int) (*User, error) {
	var user User
	query := `SELECT id, username, password, role, display_name, created_at 
	          FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
		&user.DisplayName, &user.CreatedAt,
	)

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
		// Убираем лайк
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

// === COMMENT METHODS ===

func (r *Repository) CreateComment(postID, userID int, content string) error {
	query := `INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, postID, userID, content)
	return err
}

func (r *Repository) GetCommentsByPostID(postID int) ([]Comment, error) {
	query := `
        SELECT c.id, c.post_id, c.user_id, c.content, c.created_at,
               u.id, u.username, u.display_name, u.role, u.created_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = $1
        ORDER BY c.created_at ASC
    `

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var user User
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comment.User = &user
		comments = append(comments, comment)
	}
	return comments, nil
}

func (r *Repository) DeleteComment(commentID, userID int) error {
	// Только автор комментария может удалить его
	query := `DELETE FROM comments WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, commentID, userID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("комментарий не найден или нет прав на удаление")
	}

	return nil
}

func (r *Repository) GetPostWithComments(postID int) (*Post, []Comment, error) {
	// Получаем пост
	var post Post
	var user User
	err := r.db.QueryRow(`
        SELECT p.id, p.user_id, p.content, p.slogan, p.likes, p.comments_count, p.created_at,
               u.id, u.username, u.display_name, u.role, u.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = $1
    `, postID).Scan(
		&post.ID, &post.UserID, &post.Content, &post.Slogan, &post.Likes, &post.CommentsCount, &post.CreatedAt,
		&user.ID, &user.Username, &user.DisplayName, &user.Role, &user.CreatedAt,
	)

	if err != nil {
		return nil, nil, err
	}
	post.User = &user

	// Получаем комментарии
	comments, err := r.GetCommentsByPostID(postID)
	if err != nil {
		return &post, nil, err
	}

	return &post, comments, nil
}

// === HEROES ===
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

// === АДМИН МЕТОДЫ ===
// GetStats - получение статистики
// GetStats - получение статистики
func (r *Repository) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	// Количество пользователей
	var userCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return nil, err
	}
	stats["TotalUsers"] = userCount

	// Количество постов
	var postCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	if err != nil {
		return nil, err
	}
	stats["TotalPosts"] = postCount

	// Количество комментариев
	var commentCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&commentCount)
	if err != nil {
		return nil, err
	}
	stats["TotalComments"] = commentCount

	// Количество героев
	var heroCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM heroes").Scan(&heroCount)
	if err != nil {
		return nil, err
	}
	stats["TotalHeroes"] = heroCount

	return stats, nil
}

// GetAllUsers - получение всех пользователей
func (r *Repository) GetAllUsers() ([]User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, role, display_name, created_at 
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Role,
			&user.DisplayName, &user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUserRole - обновление роли пользователя
func (r *Repository) UpdateUserRole(userID int, role string) error {
	_, err := r.db.Exec(
		"UPDATE users SET role = $1 WHERE id = $2",
		role, userID,
	)
	return err
}

// DeleteUser - удаление пользователя
func (r *Repository) DeleteUser(userID int) error {
	// Сначала удаляем связанные данные
	_, err := r.db.Exec("DELETE FROM post_likes WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("DELETE FROM comments WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("DELETE FROM posts WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	// Затем удаляем пользователя
	_, err = r.db.Exec("DELETE FROM users WHERE id = $1", userID)
	return err
}

// GetAllPostsAdmin - получение всех постов для админки

func (r *Repository) GetAllPostsAdmin() ([]Post, error) {
	rows, err := r.db.Query(`
        SELECT 
            p.id, p.user_id, p.content, p.slogan, p.likes, p.created_at,
            COALESCE(u.id, 0) as user_id,
            COALESCE(u.username, 'Удален') as username,
            COALESCE(u.display_name, '') as display_name,
            COALESCE(u.role, 'user') as role,
            COALESCE(u.created_at, NOW()) as user_created_at
        FROM posts p
        LEFT JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var user User

		// ВАЖНО: порядок полей должен соответствовать запросу выше
		err := rows.Scan(
			// Пост поля (6 полей)
			&post.ID, &post.UserID, &post.Content, &post.Slogan,
			&post.Likes, &post.CreatedAt,
			// User поля (5 полей)
			&user.ID, &user.Username, &user.DisplayName, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			log.Printf("Ошибка сканирования поста: %v", err)
			return nil, err
		}

		post.User = &user
		posts = append(posts, post)
	}

	// Проверяем ошибки после итерации
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// DeletePostAdmin - удаление поста (админская версия)
func (r *Repository) DeletePostAdmin(postID int) error {
	// Удаляем лайки
	_, err := r.db.Exec("DELETE FROM post_likes WHERE post_id = $1", postID)
	if err != nil {
		return err
	}

	// Удаляем комментарии
	_, err = r.db.Exec("DELETE FROM comments WHERE post_id = $1", postID)
	if err != nil {
		return err
	}

	// Удаляем пост
	_, err = r.db.Exec("DELETE FROM posts WHERE id = $1", postID)
	return err
}

// GetPostByID - получение поста по ID
func (r *Repository) GetPostByID(postID int) (*Post, error) {
	var post Post
	err := r.db.QueryRow(`
		SELECT id, user_id, content, slogan, likes, created_at
		FROM posts WHERE id = $1
	`, postID).Scan(
		&post.ID, &post.UserID, &post.Content,
		&post.Slogan, &post.Likes, &post.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetAllCommentsAdmin - получение всех комментариев для админки
func (r *Repository) GetAllCommentsAdmin() ([]Comment, error) {
	rows, err := r.db.Query(`
        SELECT 
            c.id, c.post_id, c.user_id, c.content, c.created_at,
            COALESCE(u.id, 0) as user_id,
            COALESCE(u.username, 'Удален') as username,
            COALESCE(u.display_name, '') as display_name,
            COALESCE(u.role, 'user') as role,
            COALESCE(u.created_at, NOW()) as user_created_at
        FROM comments c
        LEFT JOIN users u ON c.user_id = u.id
        ORDER BY c.created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var user User

		err := rows.Scan(
			// Comment поля (5 полей)
			&comment.ID, &comment.PostID, &comment.UserID,
			&comment.Content, &comment.CreatedAt,
			// User поля (5 полей)
			&user.ID, &user.Username, &user.DisplayName, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			log.Printf("Ошибка сканирования комментария: %v", err)
			return nil, err
		}

		comment.User = &user
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// DeleteCommentAdmin - удаление комментария (админская версия)
func (r *Repository) DeleteCommentAdmin(commentID int) error {
	_, err := r.db.Exec("DELETE FROM comments WHERE id = $1", commentID)
	return err
}
