package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/RobinHood3082/simplebank/internal/persistence"
	"github.com/RobinHood3082/simplebank/mail"
	"github.com/RobinHood3082/simplebank/util"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  persistence.Store
	mailer mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store persistence.Store, mailer mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					slog.Error("error processing task", "type", task.Type(), "error", err, "payload", task.Payload())
				},
			),
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user not found: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	verifyEmail, err := processor.store.CreateVerifyEmail(
		ctx,
		persistence.CreateVerifyEmailParams{
			Username:   user.Username,
			Email:      user.Email,
			SecretCode: util.RandomString(32),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	subject := "Simple Bank: Verify your email"
	verifyUrl := fmt.Sprintf("http://localhost:8080/api/v1/verify_email?email_id=%d&secret_code=%s",
		verifyEmail.ID,
		verifyEmail.SecretCode,
	)
	content := fmt.Sprintf(
		`Hello %s, <br/>
		Thank you for registering with us. <br/>
		Please <a href="%s">click here</a> to verify your email address.`,
		user.Username,
		verifyUrl,
	)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("type", "recieved task", task.Type(), "payload", task.Payload(), "user", user.Username)

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendAccountCreatedEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendAccountCreatedEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user not found: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	subject := "Simple Bank: New Account Created"
	content := fmt.Sprintf(
		`Hello %s, <br/>
		Your new account with ID: %s has been created. <br/>
		Current balance: %d %s.`,
		payload.Username,
		payload.AccountID,
		payload.Balance,
		payload.Currency,
	)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("type", "recieved task", task.Type(), "payload", task.Payload(), "user", user.Username)

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendBalanceAddedEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendBalanceAddedEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user not found: %w", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	subject := "Simple Bank: Balance Added"
	content := fmt.Sprintf(
		`Hello %s, <br/>
		Your account with ID: %s has been credited with %d %s. <br/>
		Current balance: %d %s.`,
		payload.Username,
		payload.AccountID,
		payload.AddedBalance,
		payload.Currency,
		payload.NewBalance,
		payload.Currency,
	)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("type", "recieved task", task.Type(), "payload", task.Payload(), "user", user.Username)

	return nil
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskSendAccountCreatedEmail, processor.ProcessTaskSendAccountCreatedEmail)
	mux.HandleFunc(TaskSendBalanceAddedEmail, processor.ProcessTaskSendBalanceAddedEmail)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}
