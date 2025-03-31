package utils

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

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

func BatchInsertPlayer(pls []TablePlayer) (int, error) {
	sinceConn := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(ctx)
	durationConn := time.Since(sinceConn)

	batch := &pgx.Batch{}
	sqlInsert := `insert into public.player(name, group_id) values($1, $2) on conflict (group_id, name) do nothing`
	for i := 0; i < len(pls); i++ {
		batch.Queue(sqlInsert, pls[i].Name, pls[i].Group_Id)
	}

	err = conn.SendBatch(ctx, batch).Close()
	durationBatch := time.Since(sinceConn)
	log.Println("BatchInsertPlayer", "durationConn", durationConn, "durationBatch", durationBatch)

	if err != nil {
		log.Println("batch fail", err)
		return 0, err
	}

	return 0, nil
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
