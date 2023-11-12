package utils

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
)

var DSN = "postgres://postgres:ygtianlou@localhost:5432/postgres"

type Table_pt struct {
	Id           int
	Paipu_url    string
	Player_count int
	State        int
	Desc         string
	Timestamp    string // todo dev
	Pl_1         string
	Pl_2         string
	Pl_3         string
	Pl_4         string
	Pt_1         int
	Pt_2         int
	Pt_3         int
	Pt_4         int
	Zy_1         int
	Zy_2         int
	Zy_3         int
	Zy_4         int
}

// todo query by name
func QueryPtByConds(pl string, date string) ([]Table_pt, error) {
	conn, err := pgx.Connect(context.Background(), DSN)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	return nil, nil
}

// todo insert if not exists
func InsertPt(onePt Table_pt) (error, error) {
	conn, err := pgx.Connect(context.Background(), DSN)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	return nil, nil
}

func genTable() {

}
