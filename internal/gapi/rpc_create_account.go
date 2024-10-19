package gapi

import (
	"context"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/RobinHood3082/simplebank/internal/pb"
	persistence "github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/pkg/validator"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/RobinHood3082/simplebank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	authPayload, err := server.authorizeUser(
		ctx,
		[]string{util.BankerRole, util.DepositorRole},
	)

	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateCreateAccountRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if authPayload.Role != util.BankerRole && authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot create account for another user")
	}

	arg := persistence.CreateAccountTxParams{
		CreateAccountParams: persistence.CreateAccountParams{
			Owner:    req.Username,
			Balance:  0,
			Currency: req.GetCurrency(),
		},
		AfterCreate: func(account persistence.Account) error {
			accountIDStr := strconv.FormatInt(account.ID, 10)
			if len(accountIDStr) > 4 {
				accountIDStr = accountIDStr[len(accountIDStr)-4:]
			}

			taskPayload := worker.PayloadSendAccountCreatedEmail{
				Username:  account.Owner,
				Balance:   account.Balance,
				Currency:  account.Currency,
				AccountID: "xxxx" + accountIDStr,
			}

			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.QueueCritical),
			}

			return server.taskDistributor.DistributeTask(ctx, worker.TaskSendAccountCreatedEmail, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateAccountTx(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user does not exist")
		}

		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation, pgerrcode.ForeignKeyViolation:
				return nil, status.Errorf(codes.AlreadyExists, "account creation forbidden")
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create account: %s", err)
	}

	rsp := &pb.CreateAccountResponse{
		Account: convertAccount(txResult.Account),
	}

	return rsp, nil
}

func validateCreateAccountRequest(req *pb.CreateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateCurrency(req.GetCurrency()); err != nil {
		violations = append(violations, fieldViolation("currency", err))
	}

	return violations
}
