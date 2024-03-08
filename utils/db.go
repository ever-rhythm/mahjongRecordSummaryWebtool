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

type StQueryGroupRank struct {
	Name  string
	Pt    int
	Zy    int
	Total int
	Cnt_1 int
	Cnt_2 int
	Cnt_3 int
	Cnt_4 int
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

func InsertPlayer(onePlayer TablePlayer) (int, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(context.Background())

	batch := &pgx.Batch{}
	batch.Queue(`insert into public.player(name, group_id) values($1, $2) on conflict (group_id, name) do nothing `,
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

func BatchInsertPlayer(pls []TablePlayer) (int, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(context.Background())

	batch := &pgx.Batch{}
	sqlInsert := `insert into public.player(name, group_id) values($1, $2) on conflict (group_id, name) do nothing`
	for i := 0; i < len(pls); i++ {
		batch.Queue(sqlInsert, pls[i].Name, pls[i].Group_Id)
	}

	err = conn.SendBatch(context.Background(), batch).Close()
	if err != nil {
		log.Println("batch fail", err)
		return 0, err
	}

	return 0, nil
}

func BatchInsertPaipu(paipus []TablePaipu) (int, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(context.Background())

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

	err = conn.SendBatch(context.Background(), batch).Close()
	if err != nil {
		log.Println("batch fail", err)
		return 0, err
	}

	return 0, nil
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

// todo dev
func InsertUpdateGroupRankPlayer(onePaipu TablePaipu) (int, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(context.Background())

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	sqlInsert := ""
	_, err = tx.Exec(context.Background(), sqlInsert)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return 0, err
	}

	return 0, nil
}

func QueryVipPlayer(groupId int, pls []string, dateEnd string) ([]TablePlayer, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select group_id,name from public.player_vip where name = any($1) and group_id = $2 and time_end > $3`
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

func QueryGroupRank(groupId int, dateStart string) ([]StQueryGroupRank, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return nil, err
	}
	defer conn.Close(context.Background())

	sqlQuery := `select name,pt,zy,total,cnt_1,cnt_2,cnt_3,cnt_4 from public.group_rank where group_id = $1 and time_start = $2`

	rows, err := conn.Query(context.Background(), sqlQuery, groupId, dateStart)
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[StQueryGroupRank])
	if err != nil {
		log.Println("db query fail", err)
		return nil, err
	}

	return rets, nil
}

func QueryGroupPlayerOpponentPaipu(groupId int, pl string, op string, dateStart string, dateEnd string) ([]TablePaipu, error) {
	conn, err := pgx.Connect(context.Background(), ConfigDb.Dsn)
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
		rowsPl, err := conn.Query(context.Background(), sqlQueryPls, rets[0].Player_Id)
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
