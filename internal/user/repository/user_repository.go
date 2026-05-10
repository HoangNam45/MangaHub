package repository

import (
	"errors"
	"mangahub/pkg/database"
	"mangahub/pkg/models"
)

type UserRepository struct{}

// CreateUser inserts a new user into the database
func (ur *UserRepository) CreateUser(user *models.User) error {
	if user.Username == "" || user.PasswordHash == "" {
		return errors.New("username and password hash are required")
	}

	query := `
	INSERT INTO users (id, username, password_hash, created_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := database.DB.Exec(query, user.ID, user.Username, user.PasswordHash)
	if err != nil {
		// Check if it's a unique constraint violation (username already exists)
		if err.Error() == "UNIQUE constraint failed: users.username" {
			return errors.New("username already exists")
		}
		return err
	}

	return nil
}

// GetUserByUsername retrieves a user by username
func (ur *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password_hash FROM users WHERE username = ?`

	user := &models.User{}
	err := database.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (ur *UserRepository) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, username, password_hash FROM users WHERE id = ?`

	user := &models.User{}
	err := database.DB.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}
