package pgr

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

const DB_URL = "postgres://postgres:postgres@localhost:5432/postgres"

// executes operations in a transaction and rolls back the changes
func getDb() *Pgr {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DB_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting transaction: %v\n", err)
		os.Exit(1)
	}
	pgr, err := New(tx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Pgr: %v\n", err)
		os.Exit(1)
	}
	return pgr
}

type User struct {
	Id     int64   `db:"id"`
	Name   string  `db:"name"`
	Age    int     `db:"age"`
	Movies []Movie `db:"movies"`
}

type Count struct {
	Count int `db:"count"`
}

type Movie struct {
	Id          int64   `db:"id"`
	Name        string  `db:"name"`
	Description *string `db:"description"`
}
