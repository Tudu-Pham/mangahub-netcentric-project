# MANGAHUB

MangaHub is a Go-based net-centric application for tracking manga reading progress, managing personal libraries, and demonstrating multi-protocol communication through HTTP, TCP, UDP, gRPC, and WebSocket. This project is developed for the Net-Centric Programming course under the School of Information Technology, International University – VNU.

## Struture of Code

mangahub/
├── .env
├── .gitignore
├── go.mod
├── go.sum
├── README.md
├── structure.txt
├── cmd/
│ └── server/
│ └── main.go
├── data/
│ └── manga.json
├── docs/
│ ├── api.md
│ └── demo.md
├── internal/
│ ├── auth/
│ │ ├── handler.go
│ │ ├── middleware.go
│ │ └── service.go
│ ├── grpc/
│ │ └── server.go
│ ├── manga/
│ │ ├── handler.go
│ │ ├── repository.go
│ │ └── service.go
│ ├── tcp/
│ │ └── server.go
│ ├── udp/
│ │ └── server.go
│ ├── user/
│ │ ├── handler.go
│ │ ├── repository.go
│ │ └── service.go
│ └── websocket/
│ └── server.go
├── pkg/
│ ├── database/
│ │ └── sqlite.go
│ ├── models/
│ │ ├── chat.go
│ │ ├── manga.go
│ │ ├── notification.go
│ │ ├── progress.go
│ │ └── user.go
│ └── utils/
│ ├── jwt.go
│ ├── logger.go
│ ├── password.go
│ └── response.go
├── proto/
│ ├── mangahub.pb.go
│ └── mangahub.proto
└── web/
│ ├── chat.html
│ ├── index.html
│ ├── main.js
│ └── style.css
