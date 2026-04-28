# MangaHub Demo Guide

## 1. Demo Objective

This document provides a step-by-step guide for demonstrating the **MangaHub - Manga & Comic Tracking System** project.

The main objective of the demo is to show that the system successfully implements and integrates the five required network communication protocols:

1. **HTTP REST API**: user registration, login, manga search, library management, and reading progress update.
2. **TCP**: real-time reading progress synchronization between authenticated clients.
3. **UDP**: lightweight notification registration and broadcasting.
4. **WebSocket**: real-time chat with join, chat, and leave events.
5. **gRPC**: internal service communication for manga retrieval, manga search, and progress update.

---

## 2. Demo Preparation

### 2.1. Required Tools

Before starting the demo, make sure the following tools are available:

- Go programming environment
- SQLite database
- Postman or Thunder Client for HTTP, WebSocket, and gRPC testing
- Terminal or command prompt
- `grpcurl` if testing gRPC from the terminal

Check Go installation:

```bash
go version
```

Install or verify dependencies:

```bash
go mod tidy
```

Run tests if available:

```bash
go test ./...
```

---

## 3. Start the Main Server

From the project root directory, run:

```bash
go run cmd/server/main.go
```

If the server starts successfully, the terminal should show messages similar to:

```text
TCP Sync Server running at :9090
UDP Notification Server running at :7070
gRPC server running at :50051
Server running at http://localhost:8080
```

Service addresses used in this demo:

| Service | Protocol | Address |
|---|---|---|
| HTTP REST API | HTTP | `http://localhost:8080` |
| TCP Sync Server | TCP | `localhost:9090` |
| UDP Notification Server | UDP | `localhost:7070` |
| WebSocket Chat | WebSocket | `ws://localhost:8080/ws/chat` |
| gRPC Service | gRPC | `localhost:50051` |

---

## 4. HTTP REST API Demo

### 4.1. Test Server Health

Request:

```http
GET http://localhost:8080/ping
```

Expected response:

```json
{
  "message": "pong"
}
```

Demo explanation:

> The `/ping` endpoint is used to verify that the HTTP server is running correctly.

---

### 4.2. Register a User

Request:

```http
POST http://localhost:8080/auth/register
Content-Type: application/json
```

Request body:

```json
{
  "username": "testuser",
  "email": "testuser@example.com",
  "password": "123456"
}
```

Expected response:

```json
{
  "message": "user created"
}
```

Demo explanation:

> The registration endpoint receives the username, email, and password. The password is hashed using bcrypt before being stored in the SQLite database.

---

### 4.3. Login and Get JWT Token

Request:

```http
POST http://localhost:8080/auth/login
Content-Type: application/json
```

Request body:

```json
{
  "username": "testuser",
  "password": "123456"
}
```

Expected response:

```json
{
  "token": "JWT_TOKEN_HERE"
}
```

Copy the returned token. It will be used for protected API endpoints.

Demo explanation:

> After successful login, the server generates a JWT token containing the user ID. This token is required for accessing protected endpoints.

---

### 4.4. Get Current User

Request:

```http
GET http://localhost:8080/me
Authorization: Bearer JWT_TOKEN_HERE
```

Expected response:

```json
{
  "user_id": "USER_ID_HERE"
}
```

Demo explanation:

> The `/me` endpoint proves that the authentication middleware can validate the JWT token and extract the user ID.

---

## 5. Manga API Demo

### 5.1. Get Manga List

Request:

```http
GET http://localhost:8080/manga
```

Expected response example:

```json
{
  "page": 1,
  "limit": 10,
  "results": [
    {
      "id": "one-piece",
      "title": "One Piece",
      "author": "Oda Eiichiro",
      "genres": "Action,Adventure,Shounen",
      "status": "ongoing",
      "total_chapters": 1100,
      "description": "..."
    }
  ]
}
```

Demo explanation:

> The `/manga` endpoint returns manga data stored in the database. The data can be loaded from the `data/manga.json` file.

---

### 5.2. Search Manga by Title or Author

Search by title:

```http
GET http://localhost:8080/manga?q=one
```

Search by author:

```http
GET http://localhost:8080/manga?q=oda
```

Demo explanation:

> The query parameter `q` is used to search manga by title or author using SQL `LIKE` conditions.

---

### 5.3. Filter Manga by Genre and Status

Filter by genre:

```http
GET http://localhost:8080/manga?genre=Action
```

Filter by status:

```http
GET http://localhost:8080/manga?status=ongoing
```

Combined search and filter:

```http
GET http://localhost:8080/manga?q=one&genre=Action&status=ongoing&page=1&limit=5
```

Demo explanation:

> The manga API supports filtering by genre, status, and pagination through `page` and `limit`.

---

### 5.4. Get Manga Details

Request:

```http
GET http://localhost:8080/manga/one-piece
```

Expected response:

```json
{
  "id": "one-piece",
  "title": "One Piece",
  "author": "Oda Eiichiro",
  "genres": "Action,Adventure,Shounen",
  "status": "ongoing",
  "total_chapters": 1100,
  "description": "..."
}
```

Demo explanation:

> This endpoint retrieves detailed information about a manga by its ID.

---

## 6. Library and Reading Progress Demo

### 6.1. Add Manga to User Library

Request:

```http
POST http://localhost:8080/users/library
Authorization: Bearer JWT_TOKEN_HERE
Content-Type: application/json
```

Request body:

```json
{
  "manga_id": "one-piece",
  "status": "reading",
  "current_chapter": 1
}
```

Expected response:

```json
{
  "message": "manga added to library"
}
```

Demo explanation:

> This endpoint adds a manga to the authenticated user's personal library. Since it is under `/users`, the request must include a valid JWT token.

---

### 6.2. Get User Library

Request:

```http
GET http://localhost:8080/users/library
Authorization: Bearer JWT_TOKEN_HERE
```

Expected response:

```json
[
  {
    "id": "one-piece",
    "title": "One Piece",
    "author": "Oda Eiichiro",
    "current_chapter": 1,
    "status": "reading",
    "updated_at": "..."
  }
]
```

Demo explanation:

> The server extracts the user ID from the JWT token, then joins the `user_progress` table with the `manga` table to return the user's reading library.

---

### 6.3. Update Reading Progress

Request:

```http
PUT http://localhost:8080/users/progress
Authorization: Bearer JWT_TOKEN_HERE
Content-Type: application/json
```

Request body:

```json
{
  "manga_id": "one-piece",
  "current_chapter": 10
}
```

Expected response:

```json
{
  "message": "progress updated",
  "manga_id": "one-piece",
  "current_chapter": 10
}
```

Demo explanation:

> The update progress feature validates the request before updating the database. It checks whether the manga exists, whether the chapter number is valid, whether it exceeds the total number of chapters, and whether the manga is already in the user's library. After a successful update, the system triggers TCP synchronization and UDP notification.

---

### 6.4. Validation Error Demo

#### Invalid negative chapter

Request body:

```json
{
  "manga_id": "one-piece",
  "current_chapter": -1
}
```

Expected response:

```json
{
  "error": "current_chapter must be greater than 0"
}
```

#### Manga does not exist

Request body:

```json
{
  "manga_id": "abcxyz",
  "current_chapter": 10
}
```

Expected response:

```json
{
  "error": "manga not found"
}
```

#### Manga is not in the user's library

Request body:

```json
{
  "manga_id": "naruto",
  "current_chapter": 10
}
```

Expected response:

```json
{
  "error": "manga is not in your library. Please add it first"
}
```

Demo explanation:

> These error cases show that the system does not blindly update the database. It validates both the input and the current data state before updating reading progress.

---

## 7. TCP Progress Synchronization Demo

### 7.1. Purpose

The TCP server is used for real-time reading progress synchronization. A TCP client must authenticate using a JWT token immediately after connecting. If the token is valid, the server stores the connection with the corresponding user ID. When the user updates reading progress, the TCP server sends the update only to clients belonging to the same user.

---

### 7.2. Run TCP Client

If the project has a TCP test client, run:

```bash
go run test/tcp_client.go
```

Example TCP client:

```go
package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	token := "JWT_TOKEN_HERE"

	authMsg := fmt.Sprintf(`{"type":"auth","token":"%s"}`+"\n", token)
	conn.Write([]byte(authMsg))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("FROM TCP SERVER:", scanner.Text())
	}
}
```

Expected output with a valid token:

```json
{"type":"auth_success","message":"TCP authentication successful"}
```

Expected output with an invalid token:

```json
{"type":"error","message":"invalid token"}
```

---

### 7.3. Trigger TCP Broadcast

After the TCP client is authenticated, call HTTP UpdateProgress:

```http
PUT http://localhost:8080/users/progress
Authorization: Bearer JWT_TOKEN_HERE
Content-Type: application/json
```

Request body:

```json
{
  "manga_id": "one-piece",
  "current_chapter": 11
}
```

The TCP client should receive:

```json
{
  "user_id": "USER_ID_HERE",
  "manga_id": "one-piece",
  "chapter": 11,
  "timestamp": 1710000000
}
```

Demo explanation:

> TCP synchronization uses a server-side broadcast mechanism. However, the broadcast is filtered by user ID, so progress data is sent only to the correct user's TCP clients.

---

## 8. UDP Notification Demo

### 8.1. Purpose

The UDP server is used for lightweight notifications. A UDP client sends a registration packet, the server stores the client address, and later broadcasts notifications to registered clients.

---

### 8.2. Run UDP Client

If the project has a UDP client:

```bash
go run test/udp/udp_client.go
```

Or, if using a Python client:

```bash
python test/client.py
```

The client should send a registration packet to the UDP server at `localhost:7070`.

Example registration message:

```json
{
  "type": "register"
}
```

Expected response:

```json
{
  "type": "registered",
  "message": "UDP client registered"
}
```

---

### 8.3. Trigger UDP Notification

Option 1: Call HTTP UpdateProgress. After progress is updated, the server sends a notification:

```json
{
  "type": "progress_update",
  "message": "User updated reading progress"
}
```

Option 2: If the project includes a separate chapter notification route, call:

```http
POST http://localhost:8080/notify/chapter
Content-Type: application/json
```

Request body:

```json
{
  "manga_id": "one-piece",
  "message": "New chapter released"
}
```

Demo explanation:

> UDP does not guarantee reliable delivery like TCP, but it is fast and lightweight, which makes it suitable for notification broadcasting.

---

## 9. WebSocket Chat Demo

### 9.1. Purpose

WebSocket is used for real-time chat. When a user joins, the server broadcasts a `join` event. When the user sends a message, the server broadcasts a `chat` event. When the user disconnects, the server broadcasts a `left` event.

---

### 9.2. Connect WebSocket Clients

Use Postman WebSocket request or another WebSocket client.

First, log in via HTTP (`POST /auth/login`) and copy the JWT token.

Client 1 (authenticated):

```text
ws://localhost:8080/ws/chat?token=JWT_TOKEN_USER_1
```

Client 2 (authenticated):

```text
ws://localhost:8080/ws/chat?token=JWT_TOKEN_USER_2
```

When Client 2 connects, Client 1 should receive:

```json
{
  "type": "join",
  "user_id": "USER_ID_2",
  "username": "username_from_db_2",
  "message": "username_from_db_2 joined the chat",
  "timestamp": 1710000000
}
```

---

### 9.3. Send Chat Message

Client sends:

```json
{
  "message": "hello everyone"
}
```

All connected clients should receive:

```json
{
  "type": "chat",
  "user_id": "USER_ID_1",
  "username": "username_from_db_1",
  "message": "hello everyone",
  "timestamp": 1710000000
}
```

---

### 9.4. Disconnect Client

When a client disconnects, the remaining clients should receive:

```json
{
  "type": "left",
  "user_id": "USER_ID_2",
  "username": "username_from_db_2",
  "message": "username_from_db_2 left the chat",
  "timestamp": 1710000000
}
```

Demo explanation:

> WebSocket keeps a two-way connection open between the client and server, which makes it suitable for real-time chat. The server manages active clients and broadcasts events to all connected clients.

---

## 10. gRPC Service Demo

### 10.1. Purpose

gRPC is used as an internal service communication protocol. The MangaHub gRPC service provides the following main methods:

- `GetManga`
- `SearchManga`
- `UpdateProgress`

The gRPC server runs at:

```text
localhost:50051
```

---

### 10.2. Check gRPC Service with grpcurl

List services:

```bash
grpcurl -plaintext localhost:50051 list
```

List methods:

```bash
grpcurl -plaintext localhost:50051 list mangahub.MangaService
```

Expected method list:

```text
GetManga
SearchManga
UpdateProgress
```

The service name may differ depending on the `package` definition in `proto/mangahub.proto`.

---

### 10.3. Test GetManga

```bash
grpcurl -plaintext \
  -d '{"id":"one-piece"}' \
  localhost:50051 mangahub.MangaService/GetManga
```

Expected response:

```json
{
  "id": "one-piece",
  "title": "One Piece",
  "author": "Oda Eiichiro",
  "genres": "Action,Adventure,Shounen",
  "status": "ongoing",
  "totalChapters": 1100,
  "description": "..."
}
```

---

### 10.4. Test SearchManga

```bash
grpcurl -plaintext \
  -d '{"query":"one"}' \
  localhost:50051 mangahub.MangaService/SearchManga
```

Expected response:

```json
{
  "results": [
    {
      "id": "one-piece",
      "title": "One Piece",
      "author": "Oda Eiichiro"
    }
  ]
}
```

---

### 10.5. Test UpdateProgress with gRPC

Before calling this method, the user must exist and the manga must already be in the user's library.

```bash
grpcurl -plaintext \
  -d '{"user_id":"USER_ID_HERE","manga_id":"one-piece","current_chapter":12,"status":"reading"}' \
  localhost:50051 mangahub.MangaService/UpdateProgress
```

Expected response:

```json
{
  "success": true,
  "message": "progress updated successfully",
  "userId": "USER_ID_HERE",
  "mangaId": "one-piece",
  "currentChapter": 12,
  "status": "reading"
}
```

Demo explanation:

> gRPC UpdateProgress performs logic similar to HTTP UpdateProgress. It validates the request, updates the `user_progress` table, and triggers TCP broadcast for real-time synchronization.

---

## 11. Suggested 3-5 Minute Demo Flow

### Step 1: Short Introduction

Say:

> MangaHub is a manga tracking system written in Go. It uses five communication protocols: HTTP for REST API, TCP for reading progress synchronization, UDP for notifications, WebSocket for real-time chat, and gRPC for internal services.

---

### Step 2: HTTP Auth and Manga API

1. Call `/ping`.
2. Login and obtain a JWT token.
3. Call `/me` to show token validation.
4. Call `/manga?q=one`.
5. Call `/manga/one-piece`.

---

### Step 3: Library and Progress

1. Add One Piece to the library.
2. Get the user's library.
3. Update progress to chapter 10.
4. Show the success response.

---

### Step 4: TCP Sync

1. Open a TCP client authenticated with the token.
2. Call UpdateProgress.
3. Show that the TCP client receives the real-time progress update.

---

### Step 5: UDP Notification

1. Open a UDP client.
2. Register the UDP client.
3. Trigger a notification by updating progress or calling a notification route.
4. Show that the UDP client receives the notification.

---

### Step 6: WebSocket Chat

1. Open two WebSocket clients.
2. Show the `join` event.
3. Send a chat message.
4. Disconnect one client and show the `left` event.

---

### Step 7: gRPC Service

1. Call `GetManga`.
2. Call `SearchManga`.
3. Call `UpdateProgress`.

---

## 12. Short Presentation Script

You may use the following script during the demo:

> This is MangaHub, a manga tracking system built with Go. The system allows users to register, log in, search manga, add manga to a personal library, and update their reading progress.  
>
> First, I will demonstrate the HTTP REST API. The user logs in and receives a JWT token. This token is required for protected endpoints such as `/users/library` and `/users/progress`.  
>
> Next, when the user updates reading progress, the system updates the SQLite database and sends the update to the TCP sync server. The TCP client must authenticate using a JWT token before receiving updates. The server stores each TCP connection by user ID, so progress updates are sent only to the correct user's clients.  
>
> For UDP, clients can register to receive notifications. The server can broadcast notification messages to all registered UDP clients.  
>
> For WebSocket, the system supports real-time chat. When a user joins, the server sends a `join` event. When a user sends a message, the server broadcasts a `chat` event. When the user leaves, the server sends a `left` event.  
>
> Finally, gRPC is used as an internal service. It provides `GetManga`, `SearchManga`, and `UpdateProgress`. The gRPC UpdateProgress method also updates the database and triggers TCP broadcast.  
>
> Therefore, this project integrates all five required protocols: HTTP, TCP, UDP, WebSocket, and gRPC in one cohesive application.

---

## 13. Common Demo Issues

### 13.1. 401 Unauthorized

Possible causes:

- Missing Authorization header
- Invalid token
- Expired token

Correct format:

```http
Authorization: Bearer JWT_TOKEN_HERE
```

---

### 13.2. UpdateProgress says manga is not in the library

Cause:

- The manga has not been added to the user's library yet.

Fix:

- Call `POST /users/library` before calling `PUT /users/progress`.

---

### 13.3. TCP client does not receive updates

Check:

- The TCP client sent the authentication message.
- The JWT token belongs to the same user who is updating progress.
- The HTTP UpdateProgress request succeeded.
- The TCP server is running on `:9090`.

---

### 13.4. WebSocket disconnects immediately

Common cause:

- The client sent a plain string instead of valid JSON.

Correct message format:

```json
{
  "message": "hello"
}
```

---

### 13.5. gRPC UpdateProgress method does not appear

Possible causes:

- `.pb.go` files were not regenerated.
- The server was not restarted.
- Postman did not re-import the updated `.proto` file.

Fix:

```bash
protoc --go_out=. --go-grpc_out=. proto/mangahub.proto
```

---

## 14. Demo Conclusion

End the demo with:

> MangaHub successfully demonstrates the main features of a manga tracking system and integrates five network protocols in Go. HTTP handles the main REST API, TCP synchronizes reading progress, UDP sends notifications, WebSocket supports real-time chat, and gRPC provides internal service communication. The system also includes SQLite persistence, JWT authentication, validation, and protocol-specific test clients.
