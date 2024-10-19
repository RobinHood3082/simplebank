package persistence

import (
	"context"
)

type CreateAccountTxParams struct {
	CreateAccountParams
	AfterCreate func(account Account) error
}

type CreateAccountTxResult struct {
	Account Account
}

func (store *PgStore) CreateAccountTx(ctx context.Context, arg CreateAccountTxParams) (CreateAccountTxResult, error) {
	var result CreateAccountTxResult

	err := store.execTx(
		ctx,
		func(q *Queries) error {
			var err error

			_, err = q.GetUser(ctx, arg.CreateAccountParams.Owner)
			if err != nil {
				return err
			}

			result.Account, err = q.CreateAccount(ctx, arg.CreateAccountParams)
			if err != nil {
				return err
			}

			return arg.AfterCreate(result.Account)
		},
	)

	return result, err
}
