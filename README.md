# MangaHub Net-Centric Project

MangaHub is a Go backend for tracking manga libraries and reading progress. It was built as a net-centric programming project and demonstrates how multiple communication protocols can work together in a single application.

The project combines:

- HTTP REST APIs for authentication and library management
- TCP for real-time reading progress synchronization
- UDP for lightweight notification delivery
- WebSocket for real-time chat
- gRPC for internal manga service operations
- SQLite for persistent storage

## Overview

The server starts several services at the same time:

| Service | Protocol | Default Address | Purpose |
|---|---|---|---|
| Main API | HTTP | `http://localhost:8080` | Health check, auth, manga listing, user library, progress updates |
| Chat | WebSocket | `ws://localhost:8080/ws/chat` | Real-time messaging between connected clients |
| Sync server | TCP | `localhost:9090` | Sends reading progress updates to authenticated clients |
| Notification server | UDP | `localhost:7070` | Broadcasts lightweight progress notifications |
| Manga service | gRPC | `localhost:50051` | Manga lookup, search, and progress update RPCs |

## Main Features

### HTTP REST API

- Register a user with username, email, and password
- Log in and receive a JWT token
- Check the current authenticated user with `/me`
- List manga with pagination and optional filters
- Get manga details by ID
- Seed sample manga data from `data/manga.json`
- Add manga to a user's library
- Retrieve the authenticated user's library
- Update reading progress

### TCP Progress Synchronization

- TCP clients authenticate using a JWT token
- When progress is updated, the server pushes a JSON message to connected clients of the same user

### UDP Notifications

- UDP clients can register or unregister with the notification server
- Progress updates trigger broadcast notifications to registered UDP clients

### WebSocket Chat

- Clients connect to `/ws/chat`
- Join, chat, and leave events are broadcast in real time
- `user_id` and `username` can be passed as query parameters

### gRPC Service

The `MangaService` defined in `proto/mangahub.proto` exposes:

- `GetManga`
- `SearchManga`
- `UpdateProgress`

## Tech Stack

- Go `1.25.6`
- Gin
- gRPC / Protocol Buffers
- Gorilla WebSocket
- SQLite via `modernc.org/sqlite`
- JWT authentication
- bcrypt password hashing

## Project Structure

```text
.
|-- cmd/server/main.go          # Application entry point
|-- internal/auth/              # Registration, login, JWT middleware
|-- internal/manga/             # Manga APIs and data seeding
|-- internal/user/              # Library and reading progress endpoints
|-- internal/tcp/               # TCP sync server
|-- internal/udp/               # UDP notification server
|-- internal/websocket/         # WebSocket chat server
|-- internal/grpc/              # gRPC server implementation
|-- pkg/database/               # SQLite connection and migrations
|-- proto/                      # .proto and generated gRPC code
|-- data/manga.json             # Sample manga dataset
|-- test/                       # Simple TCP/UDP/Python client utilities
|-- docs/api.md                 # Detailed API documentation
`-- docs/demo.md                # Demo script and walkthrough
```

## Database

SQLite is used as the local database. By default, the application creates and uses `mangahub.db` in the project root.

The current schema includes:

- `users`
- `manga`
- `user_progress`

## Getting Started

### Prerequisites

- Go installed
- A terminal with access to the project directory

### Install dependencies

```bash
go mod tidy
```

### Run the server

```bash
go run cmd/server/main.go
```

If startup succeeds, the application will listen on:

- HTTP: `:8080`
- TCP: `:9090`
- UDP: `:7070`
- gRPC: `:50051`

### Health check

```http
GET http://localhost:8080/ping
```

Expected response:

```json
{
  "message": "pong"
}
```

## API Summary

### Authentication

- `POST /auth/register`
- `POST /auth/login`
- `GET /me`

Protected routes require:

```http
Authorization: Bearer <JWT_TOKEN>
```

### Manga

- `GET /manga`
- `GET /manga/:id`
- `GET /seed/manga`

Supported query parameters for `GET /manga`:

- `q`
- `genre`
- `status`
- `page`
- `limit`

### User library

- `POST /users/library`
- `GET /users/library`
- `PUT /users/progress`

## Sample Workflow

1. Start the server.
2. Seed sample data with `GET /seed/manga`.
3. Register a user with `POST /auth/register`.
4. Log in with `POST /auth/login` and copy the returned JWT token.
5. Use the token to call `GET /me`.
6. Add a manga to the library with `POST /users/library`.
7. Update progress with `PUT /users/progress`.
8. Observe related messages through TCP and UDP test clients.

## Test Utilities

The repository contains small client programs for manual testing:

- `test/tcp/tcp_client.go` connects to the TCP sync server
- `test/udp/udp_client.go` registers for UDP notifications
- `test/client.py` is a simple UDP receiver script

Example:

```bash
go run test/udp/udp_client.go
```

```bash
go run test/tcp/tcp_client.go
```

## gRPC Development Notes

The protobuf definition is located at `proto/mangahub.proto`, and generated Go files already exist in `proto/`.

Implemented RPCs:

- `GetManga`
- `SearchManga`
- `UpdateProgress`

## Documentation

Additional project documents are available in:

- `docs/api.md` for detailed endpoint and protocol documentation
- `docs/demo.md` for a step-by-step demonstration guide

## Current Limitations

- The database path is currently fixed to `mangahub.db`
- The JWT signing key is hardcoded in the source
- The `.env_example` file is only a placeholder and is not used by the application
- There are no automated tests in the repository at the moment

## License

This repository does not currently include a license file.
