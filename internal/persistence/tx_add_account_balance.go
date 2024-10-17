package persistence

import (
	"context"
)

type AddAccountBalanceTxParams struct {
	AddAccountBalanceParams
	AfterCreate func(account Account) error
}

type AddAccountBalanceTxResult struct {
	Account Account
}

func (store *PgStore) AddAccountBalanceTx(ctx context.Context, arg AddAccountBalanceTxParams) (AddAccountBalanceTxResult, error) {
	var result AddAccountBalanceTxResult

	err := store.execTx(
		ctx,
		func(q *Queries) error {
			var err error

			result.Account, err = q.AddAccountBalance(ctx, AddAccountBalanceParams(arg.AddAccountBalanceParams))
			if err != nil {
				return err
			}

			return arg.AfterCreate(result.Account)
		},
	)

	return result, err
}
