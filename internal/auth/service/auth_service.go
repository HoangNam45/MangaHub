package service

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"mangahub/internal/auth/repository"
	"mangahub/pkg/models"
	"mangahub/pkg/utils"
)

type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new instance of UserService
func NewUserService() *UserService {
	return &UserService{
		repo: &repository.UserRepository{},
	}
}

// RegisterUser handles the business logic for user registration
func (us *UserService) RegisterUser(username, password string) (*models.User, error) {
	// Validate input
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	// Check if user already exists
	existingUser, _ := us.repo.GetUserByUsername(username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Create new user
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	// Save to repository
	err = us.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser handles the business logic for user authentication
func (us *UserService) LoginUser(username, password string) (string, *models.User, error) {
	// Validate input
	if username == "" || password == "" {
		return "", nil, errors.New("username and password are required")
	}

	// Get user from repository
	user, err := us.repo.GetUserByUsername(username)
	if err != nil {
		return "", nil, errors.New("invalid username or password")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", nil, errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", nil, errors.New("failed to generate token")
	}

	return token, user, nil
}
