package grpc

import (
	"context"
	"database/sql"
	"log"
	"net"

	"mangahub/internal/tcp"
	mangahubpb "mangahub/proto"

	"google.golang.org/grpc"

	pb "mangahub/proto"
)

type Server struct {
	pb.UnimplementedMangaServiceServer
	DB        *sql.DB
	TCPServer *tcp.Server
}

func NewServer(db *sql.DB, tcpServer *tcp.Server) *Server {
	return &Server{
		DB:        db,
		TCPServer: tcpServer,
	}
}

func (s *Server) GetManga(ctx context.Context, req *mangahubpb.GetMangaRequest) (*mangahubpb.MangaResponse, error) {
	var m mangahubpb.Manga

	err := s.DB.QueryRow(`
		SELECT id, title, author, genres, status, total_chapters, description
		FROM manga
		WHERE id = ?
	`, req.Id).Scan(
		&m.Id,
		&m.Title,
		&m.Author,
		&m.Genres,
		&m.Status,
		&m.TotalChapters,
		&m.Description,
	)

	if err != nil {
		return nil, err
	}

	return &mangahubpb.MangaResponse{
		Manga: &m,
	}, nil
}

func (s *Server) Start() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()

	mangahubpb.RegisterMangaServiceServer(grpcServer, s)

	log.Println("gRPC server running at :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) SearchManga(ctx context.Context, req *mangahubpb.SearchMangaRequest) (*mangahubpb.SearchMangaResponse, error) {
	rows, err := s.DB.Query(`
		SELECT id, title, author, genres, status, total_chapters, description
		FROM manga
		WHERE title LIKE ? OR author LIKE ?
	`, "%"+req.Query+"%", "%"+req.Query+"%")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*mangahubpb.Manga{}

	for rows.Next() {
		var m mangahubpb.Manga

		err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Author,
			&m.Genres,
			&m.Status,
			&m.TotalChapters,
			&m.Description,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, &m)
	}

	return &mangahubpb.SearchMangaResponse{
		Results: results,
	}, nil
}

func (s *Server) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
	userID := req.GetUserId()
	mangaID := req.GetMangaId()
	currentChapter := int(req.GetCurrentChapter())
	status := req.GetStatus()

	if userID == "" {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "user_id is required",
		}, nil
	}

	if mangaID == "" {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "manga_id is required",
		}, nil
	}

	if currentChapter <= 0 {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "current_chapter must be greater than 0",
		}, nil
	}

	if status == "" {
		status = "reading"
	}

	allowedStatuses := map[string]bool{
		"reading":      true,
		"completed":    true,
		"plan_to_read": true,
		"on_hold":      true,
		"dropped":      true,
	}

	if !allowedStatuses[status] {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "invalid status",
		}, nil
	}

	var totalChapters int
	err := s.DB.QueryRow(`
		SELECT total_chapters
		FROM manga
		WHERE id = ?
	`, mangaID).Scan(&totalChapters)

	if err == sql.ErrNoRows {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "manga not found",
		}, nil
	}

	if err != nil {
		return nil, err
	}

	if totalChapters > 0 && currentChapter > totalChapters {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "current_chapter exceeds total chapters",
		}, nil
	}

	var exists int
	err = s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM user_progress
		WHERE user_id = ? AND manga_id = ?
	`, userID, mangaID).Scan(&exists)

	if err != nil {
		return nil, err
	}

	if exists == 0 {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "manga is not in user's library. Please add it first",
		}, nil
	}

	result, err := s.DB.Exec(`
		UPDATE user_progress
		SET current_chapter = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND manga_id = ?
	`, currentChapter, status, userID, mangaID)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return &pb.UpdateProgressResponse{
			Success: false,
			Message: "progress was not updated",
		}, nil
	}

	if s.TCPServer != nil {
		s.TCPServer.BroadcastCh <- tcp.NewProgressUpdate(userID, mangaID, currentChapter)
	}

	return &pb.UpdateProgressResponse{
		Success:        true,
		Message:        "progress updated successfully",
		UserId:         userID,
		MangaId:        mangaID,
		CurrentChapter: int32(currentChapter),
		Status:         status,
	}, nil
}
