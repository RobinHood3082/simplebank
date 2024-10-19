package gapi

import (
	"context"
	"strconv"
	"time"

	"github.com/RobinHood3082/simplebank/internal/pb"
	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/RobinHood3082/simplebank/worker"
	"github.com/hibiken/asynq"
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

	res, err := server.store.AddAccountBalance(ctx, persistence.AddAccountBalanceParams{
		ID:     req.GetAccountId(),
		Amount: req.GetAmount(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add account balance")
	}

	accountIDStr := strconv.FormatInt(account.ID, 10)
	if len(accountIDStr) > 4 {
		accountIDStr = accountIDStr[len(accountIDStr)-4:]
	}

	updatedAccount, _ := server.store.GetAccount(ctx, req.GetAccountId())

	taskPayload := worker.PayloadSendBalanceAddedEmail{
		Username:     updatedAccount.Owner,
		AccountID:    "xxxx" + accountIDStr,
		AddedBalance: req.GetAmount(),
		Currency:     updatedAccount.Currency,
		NewBalance:   updatedAccount.Balance,
	}

	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}

	_ = server.taskDistributor.DistributeTask(ctx, worker.TaskSendBalanceAddedEmail, taskPayload, opts...)

	return &pb.AddAccountBalanceResponse{
		Account: convertAccount(res),
	}, nil
}
