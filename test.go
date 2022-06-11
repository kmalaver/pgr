package queryx

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

func QueryExample() {
	ctx := context.Background()
	conn, err := pgx.ConnectConfig(context.Background(), &pgx.ConnConfig{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	rows, _ := conn.Query(ctx, "SELECT * FROM users")
	for rows.Next() {
		var id int64
		var name string
		var age int
		rows.Scan(&id, &name, &age)
		rows.CommandTag().RowsAffected()

		db := New(conn)
		db.Transaction(ctx, func(ctx context.Context) error {

			_, err := db.InsertInto("table").
				Pair("name", "John").
				Pair("age", 20).
				Exec(ctx) // use transaction from context
			if err != nil {
				return err
			}
			err = CallAnotherFunction(ctx)
			if err != nil {
				return err
			}

			// nested transaction
			err = db.Transaction(ctx, func(ctx context.Context) error {
				return nil
			})
			return err
		})
	}
}

func CallAnotherFunction(ctx context.Context) error {
	return nil
}
