package gapi

import (
	"log/slog"

	"github.com/RobinHood3082/simplebank/internal/pb"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/internal/token"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/RobinHood3082/simplebank/worker"
)

// Server serves gRPC requests for our banking service
type Server struct {
	pb.UnimplementedSimpleBankServer
	store           persistence.Store
	logger          *slog.Logger
	tokenMaker      token.Maker
	config          util.Config
	taskDistributor worker.TaskDistributor
}

// NewServer creates a new gRPC server
func NewServer(store persistence.Store, logger *slog.Logger, tokenMaker token.Maker, config util.Config, taskDistributor worker.TaskDistributor) *Server {
	server := &Server{
		store:           store,
		logger:          logger,
		tokenMaker:      tokenMaker,
		config:          config,
		taskDistributor: taskDistributor,
	}
	return server
}
