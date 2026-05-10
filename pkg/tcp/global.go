package tcp

import (
	"sync"
	"time"
)

// ServerInstance holds the global TCP server instance
var (
	serverInstance *ProgressSyncServer
	instanceMu     sync.Mutex
)

// InitGlobalServer initializes the global TCP server instance
func InitGlobalServer(server *ProgressSyncServer) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	serverInstance = server
}

// GetGlobalServer returns the global TCP server instance
func GetGlobalServer() *ProgressSyncServer {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	return serverInstance
}

// BroadcastProgressUpdate sends a progress update through the global server
func BroadcastProgressUpdate(userID, mangaID string, chapter int) {
	server := GetGlobalServer()
	if server == nil {
		return
	}

	update := ProgressUpdate{
		UserID:    userID,
		MangaID:   mangaID,
		Chapter:   chapter,
		Timestamp: time.Now().Unix(),
	}

	server.BroadcastUpdate(update)
}
