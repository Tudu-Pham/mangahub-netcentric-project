# MangaHub API Documentation

This document describes the HTTP REST API and the network protocol interfaces used in the MangaHub Net-Centric Programming project.

## 1. System Overview

MangaHub is a manga tracking backend implemented in Go. The system demonstrates five network communication protocols:

| Protocol | Purpose | Address |
|---|---|---|
| HTTP REST | Authentication, manga search, library and progress management | `http://localhost:8080` |
| TCP | Real-time reading progress synchronization | `localhost:9090` |
| UDP | Lightweight notification delivery | `localhost:7070` |
| WebSocket | Real-time chat | `ws://localhost:8080/ws/chat` |
| gRPC | Internal manga service API | `localhost:50051` |

Main data tables:

| Table | Purpose |
|---|---|
| `users` | Stores user accounts and password hashes |
| `manga` | Stores manga metadata |
| `user_progress` | Stores each user's library and reading progress |

---

## 2. Running the Server

From the project root folder:

```bash
go run cmd/server/main.go
```

Expected services:

```text
HTTP server running at http://localhost:8080
TCP sync server running at :9090
UDP notification server running at :7070
gRPC server running at :50051
WebSocket endpoint available at /ws/chat
```

Health check:

```http
GET http://localhost:8080/ping
```

Response:

```json
{
  "message": "pong"
}
```

---

## 3. Authentication

Protected endpoints require this header:

```http
Authorization: Bearer <JWT_TOKEN>
```

### 3.1 Register User

```http
POST /auth/register
```

Full URL:

```text
http://localhost:8080/auth/register
```

Request body:

```json
{
  "username": "testuser",
  "email": "testuser@example.com",
  "password": "123456"
}
```

Success response:

```json
{
  "message": "user created"
}
```

Possible errors:

```json
{
  "error": "invalid input"
}
```

```json
{
  "error": "UNIQUE constraint failed: users.username"
}
```

### 3.2 Login

```http
POST /auth/login
```

Full URL:

```text
http://localhost:8080/auth/login
```

Request body:

```json
{
  "username": "testuser",
  "password": "123456"
}
```

Success response:

```json
{
  "token": "<JWT_TOKEN>"
}
```

Possible errors:

```json
{
  "error": "invalid credentials"
}
```

```json
{
  "error": "wrong password"
}
```

### 3.3 Get Current User ID

```http
GET /me
```

Full URL:

```text
http://localhost:8080/me
```

Headers:

```http
Authorization: Bearer <JWT_TOKEN>
```

Success response:

```json
{
  "user_id": "b2f1a8e4-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

---

## 4. Manga APIs

### 4.1 Seed Sample Manga Data

```http
GET /seed/manga
```

Full URL:

```text
http://localhost:8080/seed/manga
```

This endpoint inserts sample manga records into the database.

Success response:

```json
{
  "message": "manga data seeded successfully"
}
```

### 4.2 Get Manga List

```http
GET /manga
```

Full URL:

```text
http://localhost:8080/manga
```

Success response:

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
      "description": "Pirate adventure"
    }
  ]
}
```

### 4.3 Get Manga Details

```http
GET /manga/{id}
```

Example:

```text
http://localhost:8080/manga/one-piece
```

Success response:

```json
{
  "id": "one-piece",
  "title": "One Piece",
  "author": "Oda Eiichiro",
  "genres": "Action,Adventure,Shounen",
  "status": "ongoing",
  "total_chapters": 1100,
  "description": "Pirate adventure"
}
```

Not found response:

```json
{
  "error": "manga not found"
}
```

---

## 5. User Library APIs

All `/users/*` endpoints require JWT authentication.

### 5.1 Add Manga to Library

```http
POST /users/library
```

Full URL:

```text
http://localhost:8080/users/library
```

Headers:

```http
Authorization: Bearer <JWT_TOKEN>
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

Success response:

```json
{
  "message": "manga added to library"
}
```

Notes:

- This endpoint writes to the `user_progress` table.
- If the same `user_id` and `manga_id` already exist, `INSERT OR REPLACE` updates the existing entry.

### 5.2 Get User Library

```http
GET /users/library
```

Full URL:

```text
http://localhost:8080/users/library
```

Headers:

```http
Authorization: Bearer <JWT_TOKEN>
```

Success response:

```json
[
  {
    "id": "one-piece",
    "title": "One Piece",
    "author": "Oda Eiichiro",
    "current_chapter": 1,
    "status": "reading",
    "updated_at": "2026-04-28 10:30:00"
  }
]
```

### 5.3 Update Reading Progress

```http
PUT /users/progress
```

Full URL:

```text
http://localhost:8080/users/progress
```

Headers:

```http
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

Request body:

```json
{
  "manga_id": "one-piece",
  "current_chapter": 10
}
```

Success response:

```json
{
  "message": "progress updated"
}
```

Expected side effects:

1. Updates `current_chapter` and `updated_at` in the `user_progress` table.
2. Sends a TCP progress update through `BroadcastCh`.
3. Sends a UDP notification to the configured client address.

Recommended validation behavior:

- `manga_id` must exist in the `manga` table.
- `current_chapter` must be greater than `0`.
- `current_chapter` should not exceed `total_chapters` when `total_chapters` is known.
- The manga should already exist in the user's library before progress can be updated.

---

## 6. TCP Progress Sync Protocol

Address:

```text
localhost:9090
```

Purpose:

- Synchronizes reading progress updates in real time.
- The HTTP `PUT /users/progress` endpoint sends a `ProgressUpdate` into the TCP server's broadcast channel.

Progress update message format:

```json
{
  "user_id": "b2f1a8e4-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "manga_id": "one-piece",
  "chapter": 10,
  "timestamp": 1710000000
}
```

Basic TCP test client example:

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

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        fmt.Println("FROM TCP SERVER:", scanner.Text())
    }
}
```

Recommended authenticated TCP client message:

```json
{
  "type": "auth",
  "token": "<JWT_TOKEN>"
}
```

When TCP authentication is enabled, the server should:

1. Read the first JSON message from the TCP client.
2. Validate the JWT token.
3. Store the connection with the authenticated `user_id`.
4. Broadcast progress updates only to connections that belong to the same `user_id`.

---

## 7. UDP Notification Protocol

Address:

```text
localhost:7070
```

Purpose:

- Sends lightweight notifications to UDP clients.
- In the current integration, a UDP notification is sent when a user updates reading progress.

Notification message format:

```json
{
  "type": "progress_update",
  "message": "User updated reading progress"
}
```

Example Python UDP listener:

```python
import socket

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind(("127.0.0.1", 9999))

print("UDP client listening on 127.0.0.1:9999")

while True:
    data, addr = sock.recvfrom(4096)
    print("FROM UDP SERVER:", data.decode())
```

Recommended registration-based UDP design:

```json
{
  "type": "register"
}
```

Recommended unregister message:

```json
{
  "type": "unregister"
}
```

Recommended chapter notification format:

```json
{
  "type": "chapter_release",
  "manga_id": "one-piece",
  "message": "One Piece has a new chapter available"
}
```

---

## 8. WebSocket Chat API

Endpoint:

```text
ws://localhost:8080/ws/chat
```

Purpose:

- Provides real-time chat between connected clients.

### 8.1 Basic Connection

Connect using Postman WebSocket or another WebSocket client:

```text
ws://localhost:8080/ws/chat?token=<JWT_TOKEN>
```

Notes:

- The WebSocket endpoint requires a valid JWT token passed as a `token` query parameter.
- The server derives `user_id` from the token and looks up `username` from the `users` table.

Client message body:

```json
{
  "message": "hello everyone"
}
```

Broadcast response:

```json
{
  "type": "chat",
  "user_id": "USER_ID_HERE",
  "username": "username_from_db",
  "message": "hello everyone",
  "timestamp": 1710000000
}
```

### 8.2 Join / Left Events

Join event (broadcast when a client connects):

```json
{
  "type": "join",
  "user_id": "USER_ID_HERE",
  "username": "username_from_db",
  "message": "username_from_db joined the chat",
  "timestamp": 1710000000
}
```

Left event (broadcast when a client disconnects):

```json
{
  "type": "left",
  "user_id": "USER_ID_HERE",
  "username": "username_from_db",
  "message": "username_from_db left the chat",
  "timestamp": 1710000000
}
```

---

## 9. gRPC Manga Service

Address:

```text
localhost:50051
```

Proto file:

```text
proto/mangahub.proto
```

Service:

```proto
service MangaService {
  rpc GetManga(GetMangaRequest) returns (MangaResponse);
  rpc SearchManga(SearchMangaRequest) returns (SearchMangaResponse);
}
```

### 9.1 GetManga

Request:

```json
{
  "id": "one-piece"
}
```

grpcurl example:

```bash
grpcurl -plaintext \
  -d '{"id":"one-piece"}' \
  localhost:50051 mangahub.MangaService/GetManga
```

Success response:

```json
{
  "manga": {
    "id": "one-piece",
    "title": "One Piece",
    "author": "Oda Eiichiro",
    "genres": "Action,Adventure,Shounen",
    "status": "ongoing",
    "totalChapters": 1100,
    "description": "Pirate adventure"
  }
}
```

### 9.2 SearchManga

Request:

```json
{
  "query": "one"
}
```

grpcurl example:

```bash
grpcurl -plaintext \
  -d '{"query":"one"}' \
  localhost:50051 mangahub.MangaService/SearchManga
```

Success response:

```json
{
  "results": [
    {
      "id": "one-piece",
      "title": "One Piece",
      "author": "Oda Eiichiro",
      "genres": "Action,Adventure,Shounen",
      "status": "ongoing",
      "totalChapters": 1100,
      "description": "Pirate adventure"
    }
  ]
}
```

### 9.3 Recommended UpdateProgress RPC

Recommended proto definition:

```proto
rpc UpdateProgress(UpdateProgressRequest) returns (UpdateProgressResponse);

message UpdateProgressRequest {
  string user_id = 1;
  string manga_id = 2;
  int32 current_chapter = 3;
  string status = 4;
}

message UpdateProgressResponse {
  bool success = 1;
  string message = 2;
  string user_id = 3;
  string manga_id = 4;
  int32 current_chapter = 5;
  string status = 6;
}
```

Recommended request:

```json
{
  "user_id": "b2f1a8e4-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "manga_id": "one-piece",
  "current_chapter": 10,
  "status": "reading"
}
```

Recommended grpcurl command:

```bash
grpcurl -plaintext \
  -d '{"user_id":"USER_ID_HERE","manga_id":"one-piece","current_chapter":10,"status":"reading"}' \
  localhost:50051 mangahub.MangaService/UpdateProgress
```

Expected response:

```json
{
  "success": true,
  "message": "progress updated successfully",
  "userId": "USER_ID_HERE",
  "mangaId": "one-piece",
  "currentChapter": 10,
  "status": "reading"
}
```

---

## 10. Common Demo Flow

### Step 1: Start server

```bash
go run cmd/server/main.go
```

### Step 2: Check HTTP server

```http
GET http://localhost:8080/ping
```

### Step 3: Seed sample manga

```http
GET http://localhost:8080/seed/manga
```

### Step 4: Register user

```http
POST http://localhost:8080/auth/register
```

Body:

```json
{
  "username": "testuser",
  "email": "testuser@example.com",
  "password": "123456"
}
```

### Step 5: Login and copy JWT token

```http
POST http://localhost:8080/auth/login
```

Body:

```json
{
  "username": "testuser",
  "password": "123456"
}
```

### Step 6: Add manga to library

```http
POST http://localhost:8080/users/library
Authorization: Bearer <JWT_TOKEN>
```

Body:

```json
{
  "manga_id": "one-piece",
  "status": "reading",
  "current_chapter": 1
}
```

### Step 7: Start TCP client

Run a TCP client that listens on `localhost:9090`.

### Step 8: Start UDP client

Run a UDP listener on `127.0.0.1:9999`.

### Step 9: Update progress

```http
PUT http://localhost:8080/users/progress
Authorization: Bearer <JWT_TOKEN>
```

Body:

```json
{
  "manga_id": "one-piece",
  "current_chapter": 10
}
```

Expected result:

- HTTP response says progress updated.
- TCP client receives progress JSON.
- UDP client receives notification JSON.

### Step 10: Test WebSocket chat

Connect two WebSocket clients:

```text
ws://localhost:8080/ws/chat
```

Send:

```json
{
  "username": "duc",
  "message": "hello"
}
```

Both clients should receive the message.

### Step 11: Test gRPC

```bash
grpcurl -plaintext \
  -d '{"id":"one-piece"}' \
  localhost:50051 mangahub.MangaService/GetManga
```

---

## 11. Error Responses

### Missing JWT token

```json
{
  "error": "missing token"
}
```

### Invalid JWT token

```json
{
  "error": "invalid token"
}
```

### Invalid JSON body

```json
{
  "error": "invalid input"
}
```

---

## 12. HTTP Notification API (UDP Broadcast Helper)

### 12.1 Broadcast Chapter Release Notification

```http
POST /notify/chapter
```

Full URL:

```text
http://localhost:8080/notify/chapter
```

Request body:

```json
{
  "manga_id": "one-piece",
  "message": "New chapter released"
}
```

Notes:

- `manga_id` is required.
- If `message` is omitted, the server defaults it to `"New chapter released"`.

Success response:

```json
{
  "message": "chapter release notification sent",
  "manga_id": "one-piece",
  "notification": {
    "type": "chapter_release",
    "manga_id": "one-piece",
    "message": "New chapter released",
    "timestamp": 1710000000
  }
}
```

### Manga not found

```json
{
  "error": "manga not found"
}
```

---

## 12. Notes for Evaluation

This API demonstrates the required network programming concepts:

- HTTP REST for request/response application APIs.
- JWT authentication for protected user endpoints.
- TCP for real-time progress synchronization.
- UDP for lightweight notification delivery.
- WebSocket for real-time chat.
- gRPC for internal service-style manga operations.
- SQLite for persistent data storage.
