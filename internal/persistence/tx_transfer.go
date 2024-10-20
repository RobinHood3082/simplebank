package persistence

import (
	"context"
)

// TransferTxParams defines the input parameters for the transfer transaction
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult defines the output result for the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another
// It creates a transfer record, add account entries and update accounts' balance within a single database transaction
func (store *PgStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(
		ctx,
		func(q *Queries) error {
			var err error

			result.Transfer, err = q.CreateTransfer(
				ctx,
				CreateTransferParams(arg),
			)
			if err != nil {
				return err
			}

			result.FromEntry, err = q.CreateEntry(
				ctx,
				CreateEntryParams{
					AccountID: arg.FromAccountID,
					Amount:    -arg.Amount,
				},
			)
			if err != nil {
				return err
			}

			result.ToEntry, err = q.CreateEntry(
				ctx,
				CreateEntryParams{
					AccountID: arg.ToAccountID,
					Amount:    arg.Amount,
				},
			)
			if err != nil {
				return err
			}

			if arg.FromAccountID < arg.ToAccountID {
				result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)

			} else {
				result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			}
			if err != nil {
				return err
			}

			return nil
		},
	)

	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountID1, amount1, accountID2, amount2 int64) (account1, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return
	}

	return account1, account2, err
}
