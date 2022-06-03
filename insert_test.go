package queryx

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
)

type User struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func TestInsert(t *testing.T) {
	db := getDb()

	user := User{
		Name: "John",
		Age:  20,
	}
	ctx := context.Background()
	db.InsertInto("users").
		Record(user).
		Returning("id").
		Exec(ctx)
}

func getDb() Queryx {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return New(conn)
}
