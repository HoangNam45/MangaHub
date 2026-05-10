# MangaHub Technical Documentation

This document provides a comprehensive overview of the MangaHub project, including its architecture, setup instructions, and API documentation.

## Architecture Overview

MangaHub is a Go-based backend service designed to serve as a manga discovery, tracking, and communication platform. It utilizes various network protocols to provide a rich feature set:

- **REST API:** Built with the [Gin](https://gin-gonic.com/) web framework, serving as the primary interface for user authentication, manga discovery, and library management.
- **WebSocket:** Provides real-time chat functionality (`/ws/chat`), allowing users to interact with each other.
- **TCP Server:** Runs a Progress Sync Server on port `9090` for reliable, connection-oriented data synchronization.
- **UDP Server:** Runs a Notification Server on port `9091` for fast, connectionless event broadcasting.
- **gRPC:** Implements inter-service communication capabilities (seen in the `internal/grpc` module).
- **Database:** Relational database integration using GORM or standard SQL (configured in `pkg/database`) with automatic migrations on startup.

The application follows a standard Go project layout with separation of concerns:

- `cmd/`: Entry points for various servers and clients.
- `internal/`: Core business logic separated by domains (auth, manga, user, websocket, grpc).
- `pkg/`: Shared utilities and database configurations.

## Setup Instructions

### Prerequisites

- [Go](https://golang.org/doc/install) 1.20 or higher.
- A database instance (e.g., PostgreSQL, MySQL, or SQLite) configured appropriately.

### Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/HoangNam45/MangaHub
   cd mangahub
   ```

2. **Download dependencies:**

   ```bash
   go mod download
   ```

3. **Run the API Server:**
   The main API server binds to port `8080` (HTTP), `9090` (TCP), and `9091` (UDP).

   ```bash
   go run cmd/api-server/main.go
   ```

4. **Running auxiliary services (Optional):**
   You can explore other clients/servers in the `cmd/` directory:
   ```bash
   go run cmd/grpc-server/main.go
   go run cmd/grpc-client-example/main.go
   # etc.
   ```

## API Documentation

The REST API runs on `http://localhost:8080` by default.

### System

- **GET `/ping`**
  - Description: Health check endpoint.
  - Response: `{"message": "pong"}`

### Authentication

- **POST `/auth/register`**
  - Description: Register a new user.
  - Body payload: User credentials (e.g., username, email, password).
- **POST `/login`**
  - Description: Authenticate a user and receive a JWT token.

### Manga

- **GET `/manga`**
  - Description: Search or list available manga.
- **GET `/manga/:id`**
  - Description: Retrieve detailed information for a specific manga by its ID.
- **POST `/manga/notify`**
  - Description: Trigger a notification via the underlying UDP server.

### User Library (Requires Auth Token)

_All routes under `/users` require a valid JWT passed in the `Authorization` header (`Bearer <token>`)._

- **POST `/users/library`**
  - Description: Add a manga to the authenticated user's library.
- **GET `/users/library`**
  - Description: Retrieve the authenticated user's saved library.
- **PUT `/users/progress`**
  - Description: Update the user's reading progress for a specific manga.

### Chat & WebSockets

- **GET `/ws/chat`**
  - Description: Upgrade the HTTP connection to a WebSocket for real-time chat.
- **GET `/chat/stats`**
  - Description: Retrieve current chat statistics (e.g., active user count).
- **GET `/chat/history`**
  - Description: Retrieve the message history of the chat.
