package worker

const (
	TaskSendBalanceAddedEmail = "task:send_balance_added_email"
)

type PayloadSendBalanceAddedEmail struct {
	Username     string `json:"username"`
	AccountID    string `json:"account_id"`
	AddedBalance int64  `json:"added_balance"`
	Currency     string `json:"currency"`
	NewBalance      int64  `json:"new_balance"`
}
