package worker

const (
	TaskSendAccountCreatedEmail = "task:send_account_created_email"
	TaskSendVerifyEmail         = "task:send_verify_email"
)

type PayloadSendAccountCreatedEmail struct {
	Username  string `json:"username"`
	AccountID string `json:"account_id"`
	Balance   int64  `json:"balance"`
	Currency  string `json:"currency"`
}

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}
