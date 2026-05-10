package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub/internal/auth/dto"
	"mangahub/internal/auth/service"
)

type AuthController struct {
	service *service.UserService
}

func NewAuthController() *AuthController {
	return &AuthController{
		service: service.NewUserService(),
	}
}

// Register handles user registration
func (uc *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to handle registration
	user, err := uc.service.RegisterUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, dto.RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Message:  "User registered successfully",
	})
}

func (uc *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to handle login
	token, user, err := uc.service.LoginUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Return success response with token
	c.JSON(http.StatusOK, dto.LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
		Message:  "Login successful",
	})
}
