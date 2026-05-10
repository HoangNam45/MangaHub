package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingInterval = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// readLoop reads messages from the WebSocket connection
func (c *Client) readLoop() {
	defer func() {
		c.Hub.Unregister <- c.Conn
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	c.Conn.SetReadLimit(maxMessageSize)

	for {
		var msg ChatMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Validate message
		if msg.Message == "" {
			log.Printf("Empty message from %s", c.Username)
			continue
		}

		// Enforce message length limit
		if len(msg.Message) > 1000 {
			errMsg := ChatMessage{
				UserID:    "system",
				Username:  "System",
				Message:   "Message too long (max 1000 characters)",
				Timestamp: time.Now().Unix(),
				Type:      "error",
			}
			c.Send <- errMsg
			continue
		}

		// Set message metadata
		msg.UserID = c.UserID
		msg.Username = c.Username
		msg.Timestamp = time.Now().Unix()
		msg.Type = "message"

		// Broadcast the message
		c.Hub.Broadcast <- msg
	}
}

// sendLoop writes messages to the WebSocket connection
func (c *Client) sendLoop() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Start begins reading and writing for this client
func (c *Client) Start() {
	go c.readLoop()
	// sendLoop is already started in registerClient
}
