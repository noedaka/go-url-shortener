package grpc

import (
	"log"
	"net"

	"github.com/noedaka/go-url-shortener/api/proto"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/grpc/interceptor"
	"github.com/noedaka/go-url-shortener/internal/service"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	cfg     config.Config
	service service.ShortenerService
}

func NewGRPCServer(cfg config.Config, service service.ShortenerService) *GRPCServer {
	return &GRPCServer{
		cfg:     cfg,
		service: service,
	}
}

func (s *GRPCServer) StartServer() {
	listen, err := net.Listen("tcp", s.cfg.GRPCServerAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.AuthInterceptor),
	)

	handler := newHandler(s.service, s.cfg.BaseURL)

	proto.RegisterShortenerServiceServer(grpcServer, handler)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
