package gapi

import (
	"context"

	"github.com/RobinHood3082/simplebank/internal/pb"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/pkg/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(
		ctx,
		persistence.VerifyEmailTxParams{
			EmailId:    req.GetEmailId(),
			SecretCode: req.GetSecretCode(),
		},
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email: %s", err)
	}

	rsp := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}

	return rsp, nil
}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateEmailId(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}

	if err := validator.ValidateVerifyEmailCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("verify_code", err))
	}

	return violations
}
