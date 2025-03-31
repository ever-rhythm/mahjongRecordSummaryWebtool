package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

type StPlayerCareerSummary struct {
	Player_Name string
	Deal_Total  int
	Score_Total int
	Pt_Total    int
	Zy_Total    int

	Rank_1 int
	Rank_2 int
	Rank_3 int
	Rank_4 int

	Time_Index time.Time
	Time_Span  string
	Time_End   time.Time
}

type stQueryPlayerCareer struct {
	Id          int
	Deal_Total  int
	Score_Total int
	Pt_Total    int
	Zy_Total    int

	Rank_1 int
	Rank_2 int
	Rank_3 int
	Rank_4 int
}

const TIME_SPAN_TOTAL = "total"
const TIME_SPAN_DAY = "day"
const TIME_SPAN_WEEK = "week"
const TIME_SPAN_MONTH = "month"
const TIME_SPAN_YEAR = "year"

func QueryPlayerCareer(pl string, timeIndex time.Time, timeSpan string) ([]StPlayerCareerSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ConfigDb.TimeoutConnect*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, ConfigDb.Dsn)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	sqlQuery := `select player_name, deal_total, score_total, pt_total, zy_total, time_index, time_span, time_end, rank_1,rank_2,rank_3,rank_4 
       from public.player_career_summary where group_id = 0 and player_name = $1 and time_index = $2 and time_span = $3 limit 1`
	rows, err := conn.Query(ctx, sqlQuery, pl, timeIndex, timeSpan)
	if err != nil {
		return nil, err
	}

	rets, err := pgx.CollectRows(rows, pgx.RowToStructByName[StPlayerCareerSummary])
	if err != nil {
		return nil, err
	}

	return rets, nil
}

// todo test
func UpdatePlayerCareer(pl string, timeUpdate time.Time, paipuIds []int, ptVary int, zyVary int, scoreVary int, dealVary int, suffix string) error {
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

	sqlQueryCareer := `select id, deal_total, score_total, pt_total, zy_total,rank_1,rank_2,rank_3,rank_4 from public.player_career_summary where player_name = $1 and time_span = $2 limit 1`
	rowQueryCareer, err := tx.Query(ctx, sqlQueryCareer, pl, TIME_SPAN_TOTAL)
	if err != nil {
		return err
	}

	retQueryCareer, err := pgx.CollectRows(rowQueryCareer, pgx.RowToStructByName[stQueryPlayerCareer])
	if err != nil {
		return err
	}
	if len(retQueryCareer) == 0 {
		return errors.New("plcr not found" + pl)
	}

	rankSuffix := 0
	if suffix == "1" {
		rankSuffix = retQueryCareer[0].Rank_1
	} else if suffix == "2" {
		rankSuffix = retQueryCareer[0].Rank_2
	} else if suffix == "3" {
		rankSuffix = retQueryCareer[0].Rank_3
	} else if suffix == "4" {
		rankSuffix = retQueryCareer[0].Rank_4
	}

	// update total
	sqlUpdateCareer := fmt.Sprintf("update public.player_career_summary set deal_total = $1, score_total = $2, pt_total = $3, zy_total = $4 , time_end = $5 , rank_%s = $6 "+
		" where id = $7 ", suffix)
	_, err = tx.Exec(ctx, sqlUpdateCareer,
		retQueryCareer[0].Deal_Total+dealVary,
		retQueryCareer[0].Score_Total+scoreVary,
		retQueryCareer[0].Pt_Total+ptVary,
		retQueryCareer[0].Zy_Total+zyVary,
		timeUpdate,
		rankSuffix+dealVary,
		retQueryCareer[0].Id,
	)
	if err != nil {
		return err
	}

	// update month todo dev

	// update paipu status
	sqlUpdatePaipu := fmt.Sprintf("update public.paipu set status_%s = $1 where id = any($2) and status_%s = $3",
		suffix, suffix)
	_, err = tx.Exec(ctx, sqlUpdatePaipu, STATUS_DONE, paipuIds, STATUS_NEW)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
