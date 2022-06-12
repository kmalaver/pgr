package pgr

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type DML interface {
	Select(...string) *SelectBuilder
	SelectSql(query string, args ...interface{}) *SelectBuilder
	InsertInto(string) *InsertBuilder
	InsertSql(query string, args ...interface{}) *InsertBuilder
	Update(string) *UpdateBuilder
	UpdateSql(query string, args ...interface{}) *UpdateBuilder
	DeleteFrom(string) *DeleteBuilder
	DeleteSql(query string, args ...interface{}) *DeleteBuilder
	// With(name string, builder Builder) DML
}

// Conn returns the underlying pgx.Conn.
type Conn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) (err error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error)
}

type Pgr struct {
	conn   Conn
	logger Logger
}

type Config struct {
	Logger   Logger
	LogLevel LogLevel
}

// creates a new Pgr instance
func New(conn Conn, conf *Config) (*Pgr, error) {
	if conf == nil {
		conf = &Config{}
	}
	if conf.Logger == nil {
		conf.Logger = &defLogger{
			level: conf.LogLevel,
		}
	}
	return &Pgr{
		conn:   conn,
		logger: conf.Logger,
	}, nil
}

func Open(ctx context.Context, connString string, config *Config) (*Pgr, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	return New(conn, config)
}

const placeholder = "?"

func (p *Pgr) Conn() Conn {
	return p.conn
}

func (p *Pgr) exec(ctx context.Context, builder Builder) (int64, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, values := i.String(), i.Value()
	if err != nil {
		p.logger.Log(LogLevelError, kvs{
			"error": err.Error(),
			"sql":   query,
			"args":  fmt.Sprint(values),
		})
		return 0, err
	}

	if tx := getTransaction(ctx); tx != nil {
		count, err := tx.Exec(ctx, query, values...)
		return count.RowsAffected(), err
	}
	count, err := p.conn.Exec(ctx, query, values...)
	return count.RowsAffected(), err
}

func (p *Pgr) queryRows(ctx context.Context, builder Builder) (string, pgx.Rows, error) {
	i := interpolator{
		Buffer:       NewBuffer(),
		IgnoreBinary: true,
	}
	err := i.encodePlaceholder(builder, true)
	query, values := i.String(), i.Value()
	if err != nil {
		p.logger.Log(LogLevelError, kvs{
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
	rows, err := p.conn.Query(ctx, query, values...)
	return query, rows, err
}

func (p *Pgr) query(ctx context.Context, builder Builder, dest interface{}) (int, error) {
	query, rows, err := p.queryRows(ctx, builder)
	if err != nil {
		return 0, err
	}
	count, err := Load(rows, dest)
	if err != nil {
		p.logger.Log(LogLevelError, kvs{
			"error": err.Error(),
			"sql":   query,
		})
		return 0, err
	}
	return count, err
}
