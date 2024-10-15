package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

const (
	TaskSendAccountCreatedEmail = "task:send_account_created_email"
)

type PayloadSendAccountCreatedEmail struct {
	Username  string `json:"username"`
	AccountID string `json:"account_id"`
	Balance   int64  `json:"balance"`
	Currency  string `json:"currency"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendAccountCreatedEmail(
	ctx context.Context,
	payload *PayloadSendAccountCreatedEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendAccountCreatedEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	slog.Info("task enqueued", "type", info.Type, "queue", info.Queue, "payload", string(task.Payload()))

	return nil
}
