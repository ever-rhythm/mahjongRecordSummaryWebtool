package utils

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type stQueryPaipuByStatus struct {
	Id   int
	Rate string

	Pl_1 string
	Pl_2 string
	Pl_3 string
	Pl_4 string
	Pt_1 int
	Pt_2 int
	Pt_3 int
	Pt_4 int
	Zy_1 int
	Zy_2 int
	Zy_3 int
	Zy_4 int
}

type TablePaipu struct {
	Id           int
	Paipu_Url    string
	Player_Count int
	Time_Start   time.Time
	Rate         string
	Group_Id     int

	Pl_1 string
	Pl_2 string
	Pl_3 string
	Pl_4 string
	Pt_1 int
	Pt_2 int
	Pt_3 int
	Pt_4 int
	Zy_1 int
	Zy_2 int
	Zy_3 int
	Zy_4 int
}

type StTrendPaipu struct {
	Paipu_Url string

	Pl_1 string
	Pl_2 string
	Pl_3 string
	Pl_4 string
	Pt_1 int
	Pt_2 int
	Pt_3 int
	Pt_4 int
	Zy_1 int
	Zy_2 int
	Zy_3 int
	Zy_4 int
}

const STATUS_NEW = 0
const STATUS_DONE = 1

func QueryPaipuByStatus(pl string, suffix string, status int, timeStart time.Time, limit int) ([]stQueryPaipuByStatus, error) {
	sinceConn := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(ctx)
	durationConn := time.Since(sinceConn)

	sqlQuery := fmt.Sprintf("select id,rate,pl_1,pl_2,pl_3,pl_4,pt_1,pt_2,pt_3,pt_4,zy_1,zy_2,zy_3,zy_4 from public.paipu where pl_%s = $1 and status_%s = $2 and time_start >= $3 order by id asc limit $4",
		suffix, suffix)
	rows, err := conn.Query(ctx, sqlQuery, pl, status, timeStart, limit)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}
	durationQuery := time.Since(sinceConn)

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[stQueryPaipuByStatus])
	log.Println("QueryPaipuByStatus", "durationConn", durationConn, "durationQuery", durationQuery)

	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}

func QueryGroupPlayerOpponentPaipu(groupId int, pl string, op string, dateStart string, dateEnd string) ([]TablePaipu, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select * from public.paipu where group_id = $1 
				and $2 = any(array[pl_1,pl_2,pl_3,pl_4]) and $3 = any(array[pl_1,pl_2,pl_3,pl_4])
				and time_start >= $4 and time_start <= $5 order by time_start asc`
	rows, err := conn.Query(context.Background(), sqlQuery, groupId, pl, op, dateStart, dateEnd)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[TablePaipu])
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}

func QueryGroupPlayerPaipu(groupId int, pls []string, dateStart string, dateEnd string) ([]TablePaipu, error) {
	sinceConn := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(ctx)
	durationConn := time.Since(sinceConn)

	sqlQuery := `select * from public.paipu where group_id = $1 
				and ( pl_1 = any($2) or pl_2 = any($2) or pl_3 = any($2) or pl_4 = any($2) )
				and time_start >= $3 and time_start <= $4 order by time_start asc`
	rows, err := conn.Query(ctx, sqlQuery, groupId, pls, dateStart, dateEnd)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}
	durationQuery := time.Since(sinceConn)

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[TablePaipu])
	log.Println("QueryGroupPlayerPaipu", "durationConn", durationConn, "durationQuery", durationQuery)

	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}

func BatchInsertPaipu(paipus []TablePaipu) (int, error) {
	sinceConn := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err, paipus)
		return 0, err
	}
	defer conn.Close(ctx)
	durationConn := time.Since(sinceConn)

	batch := &pgx.Batch{}
	sqlInsert := `insert into public.paipu(
                     paipu_url,time_start,player_count,
					pl_1,pl_2,pl_3,pl_4,pt_1,pt_2,pt_3,pt_4,zy_1,zy_2,zy_3,zy_4,
					rate,group_id
                     ) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) on conflict (paipu_url) do nothing `

	for i := 0; i < len(paipus); i++ {
		batch.Queue(sqlInsert,
			paipus[i].Paipu_Url,
			paipus[i].Time_Start,
			paipus[i].Player_Count,
			paipus[i].Pl_1,
			paipus[i].Pl_2,
			paipus[i].Pl_3,
			paipus[i].Pl_4,
			paipus[i].Pt_1,
			paipus[i].Pt_2,
			paipus[i].Pt_3,
			paipus[i].Pt_4,
			paipus[i].Zy_1,
			paipus[i].Zy_2,
			paipus[i].Zy_3,
			paipus[i].Zy_4,
			paipus[i].Rate,
			paipus[i].Group_Id)
	}

	err = conn.SendBatch(ctx, batch).Close()
	durationBatch := time.Since(sinceConn)
	log.Println("BatchInsertPaipu", "durationConn", durationConn, "durationBatch", durationBatch)

	if err != nil {
		log.Println("batch fail", err, paipus)
		return 0, err
	}

	return 0, nil
}

func QueryPaipu(groupId int, dateStart string, dateEnd string) ([]TablePaipu, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select * from public.paipu where group_id = $1 and time_start >= $2 and time_start < $3 order by time_start asc`
	rows, err := conn.Query(ctx, sqlQuery, groupId, dateStart, dateEnd)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[TablePaipu])
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}
