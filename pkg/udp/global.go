package udp

import (
	"sync"
	"time"
)

// ServerInstance holds the global UDP server instance
var (
	serverInstance *NotificationServer
	instanceMu     sync.Mutex
)

// InitGlobalServer initializes the global UDP server instance
func InitGlobalServer(server *NotificationServer) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	serverInstance = server
}

// GetGlobalServer returns the global UDP server instance
func GetGlobalServer() *NotificationServer {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	return serverInstance
}

// SendChapterNotification sends a chapter release notification
func SendChapterNotification(mangaID, title, message string) {
	server := GetGlobalServer()
	if server == nil {
		return
	}

	notification := Notification{
		Type:      "chapter_release",
		MangaID:   mangaID,
		Title:     title,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	server.SendNotification(notification)
}

// SendMangaUpdateNotification sends a manga update notification
func SendMangaUpdateNotification(mangaID, title, message string) {
	server := GetGlobalServer()
	if server == nil {
		return
	}

	notification := Notification{
		Type:      "manga_update",
		MangaID:   mangaID,
		Title:     title,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	server.SendNotification(notification)
}
