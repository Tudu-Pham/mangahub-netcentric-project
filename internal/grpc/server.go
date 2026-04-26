package grpc

import (
	"context"
	"database/sql"
	"log"
	"net"

	mangahubpb "mangahub/proto"

	"google.golang.org/grpc"
)

type Server struct {
	mangahubpb.UnimplementedMangaServiceServer
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
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
