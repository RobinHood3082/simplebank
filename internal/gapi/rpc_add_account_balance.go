package gapi

import (
	"context"

	"github.com/RobinHood3082/simplebank/internal/pb"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) AddAccountBalance(ctx context.Context, req *pb.AddAccountBalanceRequest) (*pb.AddAccountBalanceResponse, error) {
	authPayload, err := server.authorizeUser(
		ctx,
		[]string{util.BankerRole, util.DepositorRole},
	)

	if err != nil {
		return nil, unauthenticatedError(err)
	}

	account, err := server.store.GetAccount(ctx, req.GetAccountId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "account not found")
	}

	if authPayload.Role != util.BankerRole && authPayload.Username != account.Owner {
		return nil, status.Errorf(codes.PermissionDenied, "cannot deposit to another user's account")
	}

	if req.GetAmount() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be positive")
	}

	arg := persistence.AddAccountBalanceTxParams{
		AddAccountBalanceParams: persistence.AddAccountBalanceParams{
			ID:     req.GetAccountId(),
			Amount: req.GetAmount(),
		},
		AfterCreate: func(account persistence.Account) error {
			return nil
		},
	}

	txResult, err := server.store.AddAccountBalanceTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update account balance")
	}

	return &pb.AddAccountBalanceResponse{
		Account: convertAccount(txResult.Account),
	}, nil
}
