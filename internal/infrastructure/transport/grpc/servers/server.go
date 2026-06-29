package servers

import (
	"fmt"
	"net"
	"service/internal/infrastructure/config"

	// Import generated proto files. Asumsi path ini benar.

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ServiceRegistry struct {
	// PermissionHandler *handlers.PermissionHandler
}

// GRPCServer membungkus instance grpc.Server.
type GRPCServer struct {
	server   *grpc.Server
	config   *config.ServerGRPCConfig
	registry *ServiceRegistry
}

// NewGRPCServer membuat instance server gRPC baru dan mendaftarkan semua service dari registry.
func NewGRPCServer(config *config.ServerGRPCConfig, registry *ServiceRegistry) *GRPCServer {
	srv := grpc.NewServer()

	// Daftarkan semua service yang ada di registry
	// if registry.PermissionHandler != nil {
	// 	permissionV1.RegisterRolPermissionServiceServer(srv, registry.PermissionHandler)
	// 	logger.Default().Info("Registered RolPermission gRPC service")
	// }

	// Aktifkan reflection agar bisa di-debug dengan tools seperti grpcurl
	reflection.Register(srv)

	return &GRPCServer{server: srv, config: config, registry: registry}
}

// Start menjalankan gRPC server.
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.Port, err)
	}
	return s.server.Serve(lis)
}

// Stop menghentikan gRPC server secara graceful.
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
