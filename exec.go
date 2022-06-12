package queryx

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

func exec(ctx context.Context, conn runner, builder Builder, logger Logger) (pgconn.CommandTag, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, values := i.String(), i.Value()
	if err != nil {
		logger.Log(LogLevelError, kvs{
			"error": err.Error(),
			"sql":   query,
			"args":  fmt.Sprint(values),
		})
		return nil, err
	}

	if tx := getTransaction(ctx); tx != nil {
		return tx.Exec(ctx, query, values...)
	}
	return conn.Exec(ctx, query, values...)
}

func queryRows(ctx context.Context, conn runner, builder Builder, logger Logger) (string, pgx.Rows, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, values := i.String(), i.Value()
	if err != nil {
		logger.Log(LogLevelError, kvs{
			"error": err.Error(),
			"sql":   query,
			"args":  fmt.Sprint(values),
		})
		return query, nil, err
	}
	if tx := getTransaction(ctx); tx != nil {
		rows, err := tx.Query(ctx, query, values...)
		return query, rows, err
	}
	rows, err := conn.Query(ctx, query, values...)
	return query, rows, err
}

func query(ctx context.Context, conn runner, builder Builder, dest interface{}, logger Logger) (int, error) {
	query, rows, err := queryRows(ctx, conn, builder, logger)
	if err != nil {
		return 0, err
	}
	count, err := Load(rows, dest)
	if err != nil {
		logger.Log(LogLevelError, kvs{
			"error": err.Error(),
			"sql":   query,
		})
		return 0, err
	}
	return count, err
}
