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

type TableGroup struct {
	Group_Id int
}

type TablePlayer struct {
	Name     string
	Group_Id int
}

type StQueryPlayer struct {
	Name             string
	Group_Id         int
	Player_Id        int
	Parent_Player_Id int
}

type StQueryPlayersByName struct {
	Name             string
	Group_Id         int
	Player_Id        int
	Parent_Player_Id int
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

func InsertPlayer(onePlayer TablePlayer) (int, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(context.Background())

	batch := &pgx.Batch{}
	batch.Queue(`insert into public.player(name, group_id) values($1, $2) on conflict (name) do nothing `,
		onePlayer.Name,
		onePlayer.Group_Id,
	)

	br := conn.SendBatch(context.Background(), batch)
	ct, err := br.Exec()
	if err != nil {
		log.Println("db batch fail", err, ct.RowsAffected())
		return 0, err
	}

	return int(ct.RowsAffected()), nil
}

func InsertPaipu(onePaipu TablePaipu) (int, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(context.Background())

	batch := &pgx.Batch{}
	batch.Queue(`insert into public.paipu(
                     paipu_url,time_start,player_count,
					pl_1,pl_2,pl_3,pl_4,pt_1,pt_2,pt_3,pt_4,zy_1,zy_2,zy_3,zy_4,
					rate,group_id
                     ) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) on conflict (paipu_url) do nothing `,
		onePaipu.Paipu_Url,
		onePaipu.Time_Start,
		onePaipu.Player_Count,
		onePaipu.Pl_1,
		onePaipu.Pl_2,
		onePaipu.Pl_3,
		onePaipu.Pl_4,
		onePaipu.Pt_1,
		onePaipu.Pt_2,
		onePaipu.Pt_3,
		onePaipu.Pt_4,
		onePaipu.Zy_1,
		onePaipu.Zy_2,
		onePaipu.Zy_3,
		onePaipu.Zy_4,
		onePaipu.Rate,
		onePaipu.Group_Id,
	)

	br := conn.SendBatch(context.Background(), batch)
	ct, err := br.Exec()
	if err != nil {
		log.Println("db batch fail", err, ct.RowsAffected())
		return 0, err
	}

	return int(ct.RowsAffected()), nil
}

func QueryGroupPlayerPaipu(groupId int, pls []string, dateStart string, dateEnd string) ([]TablePaipu, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select * from public.paipu where group_id = $1 
				and ( pl_1 = any($2) or pl_2 = any($2) or pl_3 = any($2) or pl_4 = any($2) )
				and time_start >= $3 and time_start <= $4 order by time_start asc`
	rows, err := conn.Query(context.Background(), sqlQuery, groupId, pls, dateStart, dateEnd)
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

func QueryPaipu(groupId int, dateStart string, dateEnd string) ([]TablePaipu, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select * from public.paipu where group_id = $1 and time_start >= $2 and time_start < $3 order by time_start asc`
	rows, err := conn.Query(context.Background(), sqlQuery, groupId, dateStart, dateEnd)
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

func QueryPlayerByNames(pls []string) ([]TablePlayer, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select name,group_id from public.player where name = any($1) limit 1`
	rows, err := conn.Query(context.Background(), sqlQuery, pls)
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

func QueryPlayersByName(pl string) ([]StQueryPlayersByName, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select name,group_id,player_id,parent_player_id from public.player where name = $1 limit 1`
	rows, err := conn.Query(context.Background(), sqlQuery, pl)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}
	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[StQueryPlayersByName])
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	if len(rets) > 0 && rets[0].Player_Id != 0 {
		sqlQueryPls := `select name,group_id,player_id,parent_player_id from public.player where parent_player_id = $1`
		rowsPl, err := conn.Query(context.Background(), sqlQueryPls, rets[0].Parent_Player_Id)
		if err != nil {
			log.Println("db query fail", err)
			return nil, err
		}

		retsPl, err := pgx.CollectRows(rowsPl, pgx.RowToStructByName[StQueryPlayersByName])
		if err != nil {
			log.Println("db query fail", err)
			return nil, err
		}

		return retsPl, nil
	} else {
		return rets, nil
	}
}

func QueryPlayer(groupId int) ([]StQueryPlayer, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select name,group_id,player_id,parent_player_id from public.player where group_id = $1`
	rows, err := conn.Query(context.Background(), sqlQuery, groupId)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[StQueryPlayer])
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}

func QueryGroup(code string) ([]TableGroup, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select group_id from public.group where code = $1 limit 1`
	rows, err := conn.Query(context.Background(), sqlQuery, code)
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
