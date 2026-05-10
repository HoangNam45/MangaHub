package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mangahub/internal/manga/dto"
	"mangahub/internal/manga/service"
	"mangahub/pkg/udp"
)

type MangaController struct {
	service *service.MangaService
}

// NewMangaController creates a new instance of MangaController
func NewMangaController() *MangaController {
	return &MangaController{
		service: service.NewMangaService(),
	}
}

// SearchManga handles manga search with filters
func (mc *MangaController) SearchManga(c *gin.Context) {
	var req dto.SearchMangaRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default pagination values
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Call service to search manga
	mangas, err := mc.service.SearchManga(req.Title, req.Author, req.Genre, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply pagination
	start := req.Offset
	end := start + req.Limit
	if start > len(mangas) {
		start = len(mangas)
	}
	if end > len(mangas) {
		end = len(mangas)
	}

	paginatedMangas := mangas[start:end]

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"total": len(mangas),
		"limit": req.Limit,
		"offset": req.Offset,
		"data": paginatedMangas,
	})
}

// GetMangaByID handles getting manga details by ID
func (mc *MangaController) GetMangaByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga ID is required"})
		return
	}

	// Call service to get manga
	manga, err := mc.service.GetMangaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, manga)
}

// SendNotification sends a notification about a manga event (e.g., new chapter)
// POST /manga/notify (admin endpoint)
func (mc *MangaController) SendNotification(c *gin.Context) {
	var req dto.NotificationRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default notification type
	if req.Type == "" {
		req.Type = "chapter_release"
	}

	// Get UDP server instance
	udpServer := udp.GetGlobalServer()
	if udpServer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Notification service unavailable"})
		return
	}

	// Send notification
	udpServer.SendNotification(udp.Notification{
		Type:      req.Type,
		MangaID:   req.MangaID,
		Title:     req.Title,
		Message:   req.Message,
		Timestamp: time.Now().Unix(),
	})

	// Return response
	c.JSON(http.StatusOK, dto.NotificationResponse{
		Message:        "Notification sent successfully",
		MangaID:        req.MangaID,
		ClientsNotified: udpServer.GetRegisteredClients(),
		Timestamp:      time.Now().Format(time.RFC3339),
	})
}