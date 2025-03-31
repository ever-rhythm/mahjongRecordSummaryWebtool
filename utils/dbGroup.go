package utils

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type TableGroup struct {
	Group_Id int
}

func QueryGroup(code string) ([]TableGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(ctx)

	sqlQuery := `select group_id from public.group where code = $1 limit 1`
	rows, err := conn.Query(ctx, sqlQuery, code)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[TableGroup])
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}
