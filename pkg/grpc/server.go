package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
	gGRPC "google.golang.org/grpc"
)

// Server is the GRPC server itself.
type Server struct {
	logger      *zap.Logger
	port        uint
	server      *gGRPC.Server
	registrator func(server gGRPC.ServiceRegistrar)
	listener    net.Listener
}

// NewServer is the gRPC Server constructor.
func NewServer(
	logger *zap.Logger,
	port uint,
	listener net.Listener,
	registrator func(server gGRPC.ServiceRegistrar),
) (*Server, error) {
	if port <= 80 {
		return nil, errors.New("the GRPC server port should be greater than or equal to 80")
	}

	if registrator == nil {
		return nil, errors.New("missing required dependency: Registrator")
	}

	grpcServer := gGRPC.NewServer()

	return &Server{
		logger:      logger,
		port:        port,
		registrator: registrator,
		listener:    listener,
		server:      grpcServer,
	}, nil
}

// MustNewServer is the GRPC Server constructor.
// It panics if any error is found.
func MustNewServer(
	logger *zap.Logger,
	port uint,
	listener net.Listener,
	registrator func(server gGRPC.ServiceRegistrar),
) *Server {
	server, err := NewServer(
		logger,
		port,
		listener,
		registrator,
	)
	if err != nil {
		panic(err)
	}

	return server
}

func (s Server) Run(ctx context.Context) error {
	s.registrator(s.server)

	address := fmt.Sprintf(":%d", s.port)

	listener := s.listener
	if listener == nil {
		l, err := net.Listen("tcp", address)
		if err != nil {
			s.logger.Error("failed to start to listen TCP address", zap.Error(err))
			return err
		}

		listener = l
	}

	s.logger.Info("gRPC server started:", zap.Uint("port", s.port))

	if err := s.server.Serve(listener); err != nil {
		s.logger.Error("failed to serve gRPC server", zap.Error(err))
		return err
	}

	return nil
}
