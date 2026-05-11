# MangaHub — Architecture Flow

This document describes how the MangaHub net-centric backend is structured at runtime and how traffic moves between protocols, handlers, and storage.

## 1. System context

One **Go process** hosts multiple listeners. Clients may use HTTP (REST + WebSocket), TCP, UDP, or gRPC depending on the feature. All business data that survives restarts goes through **SQLite** (`mangahub.db` by default).

```mermaid
flowchart LR
  subgraph clients [Clients]
    Web[Web / REST client]
    WSC[WebSocket chat client]
    TCPC[TCP sync client]
    UDPC[UDP notification client]
    GRPCC[gRPC client]
  end

  subgraph mangahub [MangaHub process]
    HTTP[HTTP :8080]
    TCP[TCP :9090]
    UDP[UDP :7070]
    GRPC[gRPC :50051]
    DB[(SQLite)]
  end

  Web --> HTTP
  WSC --> HTTP
  TCPC --> TCP
  UDPC --> UDP
  GRPCC --> GRPC
  HTTP --> DB
  GRPC --> DB
```

## 2. Internal components

Handlers under `internal/` are wired in `cmd/server/main.go`. Shared dependencies are the database handle and, where needed, the TCP and UDP server instances for fan-out after progress or notification events.

```mermaid
flowchart TB
  subgraph entry [Entry]
    Main["cmd/server/main.go"]
  end

  subgraph persistence [Persistence]
    DBPkg["pkg/database"]
    DB[(SQLite)]
  end

  subgraph http_layer [HTTP :8080 — Gin]
    Gin[Gin engine]
    Manga["internal/manga"]
    Auth["internal/auth"]
    User["internal/user"]
    Notify["internal/notification"]
    WS["internal/websocket"]
  end

  subgraph sidecars [Concurrent servers]
    TCP["internal/tcp :9090"]
    UDP["internal/udp :7070"]
    GRPC["internal/grpc :50051"]
  end

  Main --> DBPkg
  DBPkg --> DB
  Main --> Gin
  Main --> TCP
  Main --> UDP
  Main --> GRPC
  Main --> WS

  Gin --> Manga
  Gin --> Auth
  Gin --> User
  Gin --> Notify
  WS --> Gin

  Manga --> DB
  Auth --> DB
  User --> DB
  WS --> DB

  User --> TCP
  User --> UDP
  Notify --> UDP
  GRPC --> DB
  GRPC --> TCP
```

## 3. Startup sequence

At boot, the process prepares storage first, then starts protocol servers so they are ready before or while HTTP routes are registered.

```mermaid
sequenceDiagram
  participant Main as main.go
  participant DB as database
  participant Gin as Gin
  participant TCP as TCP server
  participant WS as WebSocket hub
  participant UDP as UDP server
  participant GRPC as gRPC server

  Main->>DB: Connect + Migrate
  Main->>Gin: New engine, CORS, routes
  Main->>TCP: go Start :9090
  Main->>WS: RegisterRoutes + go Run
  Main->>UDP: go Start :7070
  Main->>GRPC: go Start :50051
  Main->>Gin: Run :8080 blocks
```

## 4. Typical data flows

### 4.1 REST: authenticated user updates reading progress

`PUT /users/progress` persists to SQLite, then notifies real-time channels.

```mermaid
sequenceDiagram
  participant Client
  participant Gin
  participant MW as Auth middleware
  participant User as user.Handler
  participant DB as SQLite
  participant TCP as tcp.Server
  participant UDP as udp.Server

  Client->>Gin: PUT /users/progress + Bearer JWT
  Gin->>MW: AuthMiddleware
  MW->>User: UpdateProgress
  User->>DB: UPDATE user_progress
  User->>TCP: BroadcastCh progress JSON
  User->>UDP: Broadcast notification
  User-->>Client: HTTP 200
```

### 4.2 gRPC: `UpdateProgress`

The gRPC service updates the database and pushes to the same TCP sync channel (no UDP path in this RPC unless extended elsewhere).

```mermaid
sequenceDiagram
  participant Client as gRPC client
  participant S as grpc.Server
  participant DB as SQLite
  participant TCP as tcp.Server

  Client->>S: UpdateProgress
  S->>DB: persist progress
  S->>TCP: BroadcastCh progress JSON
  S-->>Client: response
```

### 4.3 HTTP: chapter release notification (UDP fan-out)

`POST /notify/chapter` does not require the full user progress pipeline; it broadcasts to registered UDP listeners.

```mermaid
sequenceDiagram
  participant Client
  participant Gin
  participant N as notification.Handler
  participant UDP as udp.Server
  participant UDPC as UDP clients

  Client->>Gin: POST /notify/chapter
  Gin->>N: NotifyChapterRelease
  N->>UDP: Broadcast
  UDP-->>UDPC: datagrams
  N-->>Client: HTTP 200
```

### 4.4 WebSocket chat

The WebSocket server shares the Gin router; clients authenticate with a JWT (for example via query string as documented in the README). Messages are relayed between connected users through the in-memory hub while user identity is resolved using the database.

```mermaid
sequenceDiagram
  participant Browser
  participant Gin as Gin :8080
  participant WS as websocket.Server
  participant Hub as Hub / goroutines
  participant DB as SQLite

  Browser->>Gin: GET /ws/chat?token=...
  Gin->>WS: upgrade connection
  WS->>DB: resolve user from token
  Browser->>Hub: join / chat / leave
  Hub-->>Browser: broadcast events
```

### 4.5 TCP progress sync (outbound from server)

TCP clients connect to `:9090`, authenticate with JWT, and receive JSON pushed when their user’s progress is updated over REST or gRPC.

```mermaid
flowchart LR
  REST["HTTP PUT /users/progress"]
  GRPC["gRPC UpdateProgress"]
  Q["tcp.Server BroadcastCh"]
  Workers["TCP writer goroutines"]
  Clients["Authenticated TCP clients"]

  REST --> Q
  GRPC --> Q
  Q --> Workers
  Workers --> Clients
```

## 5. Protocol summary

| Listener | Package | Primary responsibility |
|----------|---------|-------------------------|
| `:8080` | Gin + `internal/*` routes + `internal/websocket` | REST API, health, chat upgrade |
| `:9090` | `internal/tcp` | Push reading progress to connected TCP clients |
| `:7070` | `internal/udp` | Lightweight broadcast notifications |
| `:50051` | `internal/grpc` | Manga lookup, search, progress RPCs |

## 6. Source layout (reference)

| Path | Role |
|------|------|
| `cmd/server/main.go` | Composition root: DB, Gin, TCP, UDP, gRPC, WebSocket |
| `internal/auth` | JWT issuance and middleware |
| `internal/manga` | Manga REST + seed from `data/manga.json` |
| `internal/user` | Library + progress; bridges to TCP and UDP |
| `internal/notification` | HTTP trigger → UDP broadcast |
| `internal/tcp` | Progress sync server |
| `internal/udp` | Notification server |
| `internal/websocket` | Real-time chat |
| `internal/grpc` | `MangaService` implementation |
| `pkg/database` | SQLite connection and migrations |
| `proto/` | Contract and generated stubs |

---

For endpoint-level detail, see [api.md](./api.md). For a scripted walkthrough, see [demo.md](./demo.md).
