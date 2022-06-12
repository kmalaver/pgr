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
	runner
}

type runner interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	BeginFunc(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type queryx struct {
	runner
	logger Logger
}

type Config struct {
	Logger   Logger
	LogLevel LogLevel
}

// creates a new Queryx instance
func New(runner runner, conf *Config) (Queryx, error) {
	if conf == nil {
		conf = &Config{}
	}
	if conf.Logger == nil {
		conf.Logger = &defLogger{
			level: conf.LogLevel,
		}
	}
	return &queryx{
		runner: runner,
		logger: conf.Logger,
	}, nil
}

func Open(ctx context.Context, connString string, config *Config) (Queryx, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	return New(conn, config)
}

const placeholder = "?"
