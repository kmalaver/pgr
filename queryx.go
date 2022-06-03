package queryx

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type Queryx interface {
	Select(...string) *SelectBuilder
	SelectSql(query string, args ...interface{}) *SelectBuilder
	InsertInto(string) *InsertBuilder
	InsertSql(query string, args ...interface{}) *InsertBuilder
	Update(string) *UpdateBuilder
	UpdateSql(query string, args ...interface{}) *UpdateBuilder
	DeleteFrom(string) *DeleteBuilder
	DeleteSql(query string, args ...interface{}) *DeleteBuilder
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type runner interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	BeginFunc(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type queryx struct {
	conn runner
}

// creates a new Queryx instance
func New(conn *pgx.Conn) Queryx {
	return &queryx{conn: conn}
}

// Creates a new transactionally scoped Queryx instance
// It will execute all in a transaction and rollback on close
// Use for testing proposes
func NewTransactional(conn *pgx.Conn) (Queryx, error) {
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	return &queryx{conn: tx}, nil
}

const placeholder = "?"

func exec(ctx context.Context, conn runner, builder Builder) (pgconn.CommandTag, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, values := i.String(), i.Value()
	if err != nil {
		// TODO: send error
		return nil, err
	}

	if tx := getTransaction(ctx); tx != nil {
		return tx.Exec(ctx, query, values...)
	}
	return conn.Exec(ctx, query, values...)
}

func queryRows(ctx context.Context, conn runner, builder Builder) (pgx.Rows, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, values := i.String(), i.Value()
	if err != nil {
		// TODO: send error
		return nil, err
	}
	if tx := getTransaction(ctx); tx != nil {
		return tx.Query(ctx, query, values...)
	}
	return conn.Query(ctx, query, values...)
}

func query(ctx context.Context, conn runner, builder Builder, dest interface{}) (int, error) {

	rows, err := queryRows(ctx, conn, builder)
	if err != nil {
		return 0, err
	}
	count, err := Load(rows, dest)
	if err != nil {
		return 0, err
	}
	return count, nil
}
