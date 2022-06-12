package queryx

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type contextKey string

const transactionKey contextKey = "transaction"

func (db *queryx) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return db.BeginFunc(ctx, func(tx pgx.Tx) error {
		ctx = setTransaction(ctx, tx)
		return fn(ctx)
	})
}

func getTransaction(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(transactionKey).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func setTransaction(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, transactionKey, tx)
}
