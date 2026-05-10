package main

import (
	"fmt"

	"mangahub/internal/auth/controller"
	mangaController "mangahub/internal/manga/controller"
	"mangahub/internal/middleware"
	userController "mangahub/internal/user/controller"
	"mangahub/internal/websocket"
	"mangahub/pkg/database"
	"mangahub/pkg/tcp"
	"mangahub/pkg/udp"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()
	database.Migrate()

	// Initialize TCP server
	tcpServer := tcp.NewProgressSyncServer("9090")
	err := tcpServer.Start()
	if err != nil {
		fmt.Printf("Warning: TCP server failed to start: %v\n", err)
	} else {
		tcp.InitGlobalServer(tcpServer)
	}

	// Initialize UDP notification server
	udpServer := udp.NewNotificationServer("9091")
	err = udpServer.Start()
	if err != nil {
		fmt.Printf("Warning: UDP server failed to start: %v\n", err)
	} else {
		udp.InitGlobalServer(udpServer)
	}
	
	// Initialize WebSocket chat hub
	chatHub := websocket.NewChatHub()
	go chatHub.Run()
	
	r := gin.Default()

	// Initialize controllers
	authController := controller.NewAuthController()
	manga := mangaController.NewMangaController()
	user := userController.NewUserController()

	// Health check endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// User routes
	r.POST("/auth/register", authController.Register)
	r.POST("/login", authController.Login)

	// Manga routes
	r.GET("/manga", manga.SearchManga)
	r.GET("/manga/:id", manga.GetMangaByID)
	r.POST("/manga/notify", manga.SendNotification)

	// User library routes (protected with auth middleware)
	userRoutes := r.Group("/users")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		userRoutes.POST("/library", user.AddToLibrary)
		userRoutes.GET("/library", user.GetLibrary)
		userRoutes.PUT("/progress", user.UpdateProgress)
	}

	// WebSocket chat routes
	r.GET("/ws/chat", func(c *gin.Context) {
		chatHub.HandleWebSocket(c.Writer, c.Request)
	})
	r.GET("/chat/stats", func(c *gin.Context) {
		chatHub.HandleStats(c.Writer, c.Request)
	})
	r.GET("/chat/history", func(c *gin.Context) {
		chatHub.HandleHistory(c.Writer, c.Request)
	})

	r.Run(":8080")

	// Graceful shutdown of TCP server
	if tcpServer != nil {
		tcpServer.Stop()
	}

	// Graceful shutdown of UDP server
	if udpServer != nil {
		udpServer.Stop()
	}
}