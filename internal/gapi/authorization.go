package gapi

import (
	"context"
	"errors"
	"strings"

	"github.com/RobinHood3082/simplebank/internal/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader     = "authorization"
	authorizationTypeBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("metadata is not provided")
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return nil, errors.New("authorization header is not provided")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, errors.New("invalid authorization header format")
	}

	authorizationType := strings.ToLower(fields[0])
	if authorizationType != authorizationTypeBearer {
		return nil, errors.New("unsupported authorization type")
	}

	accessToken := fields[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, errors.New("invalid access token")
	}

	return payload, nil
}
