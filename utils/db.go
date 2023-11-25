package utils

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type configDb struct {
	Dsn string
}

var ConfigDb = configDb{}

type Table_pt struct {
	Id           int
	Paipu_url    string
	Player_count int
	//State        int
	//Desc         string
	Time_start time.Time
	Pl_1       string
	Pl_2       string
	Pl_3       string
	Pl_4       string
	Pt_1       int
	Pt_2       int
	Pt_3       int
	Pt_4       int
	Zy_1       int
	Zy_2       int
	Zy_3       int
	Zy_4       int
}

// collect 要求字段1个不少
func QueryPtByConds(pl string, timeStart int, timeEnd int) ([]Table_pt, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	//sqlQuery := `select * from pt where time_start >= to_timestamp($1) and time_start <= to_timestamp($2)`
	//rows, err := conn.Query(context.Background(), sqlQuery, timeStart, timeEnd)

	sqlQuery := `select * from pt where (Pl_1 = $1 or Pl_2 = $1 or Pl_3 = $1 or Pl_4 = $1) order by time_start asc`
	rows, err := conn.Query(context.Background(), sqlQuery, pl)
	if err != nil {
		log.Println("query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[Table_pt])
	if err != nil {
		log.Println("collect fail", err)
		return nil, err
	}

	return rets, nil
}

func QueryPts(timeStart int, timeEnd int) ([]Table_pt, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select * from pt`
	rows, err := conn.Query(context.Background(), sqlQuery)
	if err != nil {
		log.Println("query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[Table_pt])
	if err != nil {
		log.Println("collect fail", err)
		return nil, err
	}

	return rets, nil
}

// todo test
func QueryPtByUrl(paipu_url string) (error, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select count(*) as num from pt where paipu_url = $1`
	rows, err := conn.Query(context.Background(), sqlQuery, paipu_url)
	if err != nil {
		log.Println("query fail", err)
		return nil, err
	}

	log.Println(rows)

	return nil, nil
}

// insert if not exists
func InsertPt(onePt Table_pt) (error, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	batch := &pgx.Batch{}
	batch.Queue(`insert into pt(
                     paipu_url, 
                     time_start,
					player_count,pl_1,pl_2,pl_3,pl_4,pt_1,pt_2,pt_3,pt_4,zy_1,zy_2,zy_3,zy_4
                     ) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) on conflict (paipu_url) do nothing `,
		onePt.Paipu_url,
		onePt.Time_start,
		onePt.Player_count,
		onePt.Pl_1,
		onePt.Pl_2,
		onePt.Pl_3,
		onePt.Pl_4,
		onePt.Pt_1,
		onePt.Pt_2,
		onePt.Pt_3,
		onePt.Pt_4,
		onePt.Zy_1,
		onePt.Zy_2,
		onePt.Zy_3,
		onePt.Zy_4,
	)

	br := conn.SendBatch(context.Background(), batch)
	ct, err := br.Exec()

	//ct, err := conn.Exec(context.Background(), "")

	if err != nil {
		log.Println("db batch fail", err, ct.RowsAffected())
		return nil, err
	}

	return nil, nil
}

func genTable() {
	/*
		batch.Queue(`insert into pt(
		                     paipu_url,
		                     time_start,
		                     player_count,pl_1,pl_2,pl_3,pl_4,pt_1,pt_2,pt_3,pt_4,zy_1,zy_2,zy_3,zy_4
		                     ) values("$1","$2", $3, "$4", "$5", "$6", "$7", $8, $9, $10, $11, $12, $13, $14, $15) on conflict (paipu_url) do nothing `,
				onePt.Paipu_url,
				onePt.Time_start,
				onePt.Player_count,
				onePt.Pl_1,
				onePt.Pl_2,
				onePt.Pl_3,
				onePt.Pl_4,
				onePt.Pt_1,
				onePt.Pt_2,
				onePt.Pt_3,
				onePt.Pt_4,
				onePt.Zy_1,
				onePt.Zy_2,
				onePt.Zy_3,
				onePt.Zy_4,
			)

			rows := [][]any{
				{
					onePt.Paipu_url,
					onePt.Time_start,
					onePt.Player_count,
					onePt.Pl_1,
					onePt.Pl_2,
					onePt.Pl_3,
					onePt.Pl_4,
					onePt.Pt_1,
					onePt.Pt_2,
					onePt.Pt_3,
					onePt.Pt_4,
				},
			}

			ct, err := conn.CopyFrom(
				context.Background(),
				pgx.Identifier{"table_pt"},
				[]string{
					"paipu_url",
					"timestamp",
					"player_count",
					"pl_1",
					"pl_2",
					"pl_3",
					"pl_4",
					"pt_1",
					"pt_2",
					"pt_3",
					"pt_4",
				},
				pgx.CopyFromRows(rows),
			)
			if err != nil {
				log.Println("db copyfrom fail", err)
				return nil, err
			}
	*/
}
