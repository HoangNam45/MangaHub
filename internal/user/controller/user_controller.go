package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mangahub/internal/user/dto"
	"mangahub/internal/user/service"
	"mangahub/pkg/tcp"
)

type UserController struct {
	libraryService *service.UserLibraryService
}

// NewUserController creates a new instance of UserController
func NewUserController() *UserController {
	return &UserController{
		libraryService: service.NewUserLibraryService(),
	}
}

// AddToLibrary adds a manga to the user's library
// POST /users/library
func (uc *UserController) AddToLibrary(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.AddToLibraryRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to add to library
	response, err := uc.libraryService.AddToLibrary(userID.(string), req.MangaID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, response)
}

// GetLibrary retrieves the user's library
// GET /users/library
func (uc *UserController) GetLibrary(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Call service to get library
	response, err := uc.libraryService.GetUserLibrary(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, response)
}

// UpdateProgress updates the user's reading progress
// PUT /users/progress
func (uc *UserController) UpdateProgress(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateProgressRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service to update progress
	response, err := uc.libraryService.UpdateReadingProgress(userID.(string), req.MangaID, req.CurrentChapter)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Broadcast progress update to TCP clients
	tcp.BroadcastProgressUpdate(userID.(string), req.MangaID, req.CurrentChapter)

	// Return success response
	c.JSON(http.StatusOK, response)
}
