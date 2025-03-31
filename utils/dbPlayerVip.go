package utils

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type stQueryVip struct {
	Name string
	Code string
}

type stQueryVips struct {
	Name string
}

type stQueryVipCode struct {
	Name string
}

func QueryVip(pl string) ([]stQueryVip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	sqlQuery := `select name, code from public.player_vip where name = $1 limit 1`
	rows, err := conn.Query(ctx, sqlQuery, pl)
	if err != nil {
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[stQueryVip])
	if err != nil {
		return nil, err
	}

	return rets, nil
}

func QueryVipCode(code string, pl string) ([]stQueryVipCode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	sqlQuery := `select name from public.player_vip where code = $1 and name = $2 limit 1`
	rows, err := conn.Query(context.Background(), sqlQuery, code, pl)
	if err != nil {
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[stQueryVipCode])
	if err != nil {
		return nil, err
	}

	return rets, nil
}

func QueryVips() ([]stQueryVips, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	sqlQuery := `select name from public.player_vip`
	rows, err := conn.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[stQueryVips])
	if err != nil {
		return nil, err
	}

	return rets, nil
}

func QueryVipPlayer(groupId int, pls []string, dateEnd string) ([]TablePlayer, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select group_id,name from public.player_vip where name = any($1) and group_id = $2 and time_end > $3 limit 1`
	rows, err := conn.Query(context.Background(), sqlQuery, pls, groupId, dateEnd)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[TablePlayer])
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}

// todo dev
func InsertVipAndCareer(pl string, code string, timeIndex time.Time, timeNow time.Time, timeEnd time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// add vip
	sqlInsertVip := `insert into public.player_vip (name, code, time_start, time_end) values ($1, $2, $3, $4)`
	_, err = tx.Exec(ctx, sqlInsertVip, pl, code, timeNow, timeEnd)
	if err != nil {
		return err
	}

	// add total career
	sqlInsertCareerTotal := `insert into public.player_career_summary (player_name, time_span, time_index, time_start, time_end) values ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, sqlInsertCareerTotal, pl, TIME_SPAN_TOTAL, timeIndex, timeNow, timeNow)
	if err != nil {
		return err
	}

	// add month career
	sqlInsertMonth := `insert into public.player_career_summary (player_name, time_span, time_index, time_start, time_end) values ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, sqlInsertMonth, pl, TIME_SPAN_MONTH, timeIndex, timeNow, timeNow)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
