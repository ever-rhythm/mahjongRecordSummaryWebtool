package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

func GetGroupPlayerOpponentTrend(code string, date string, pl string, op string) ([]utils.TablePaipu, []string, error) {
	// query code
	retGroup, err := utils.QueryGroup(code)
	if err != nil {
		log.Println("QueryGroup fail")
		return nil, nil, err
	}

	if len(retGroup) == 0 {
		return nil, nil, errors.New("invalid code")
	}

	dateEnd, err := utils.GetNextMonthDate(date)
	if err != nil {
		log.Println("next month date invalid")
		return nil, nil, err
	}

	// query paipu
	retPaipu, err := utils.QueryGroupPlayerOpponentPaipu(retGroup[0].Group_Id, pl, op, date, dateEnd)
	if err != nil {
		log.Println("QueryGroupPlayerPaipu fail")
		return nil, nil, err
	}

	return retPaipu, []string{pl}, nil
}

func GetGroupPlayerTrend(code string, pl string, date string, half string) ([]utils.TablePaipu, []string, string, string, error) {

	// query code
	retGroup, err := utils.QueryGroup(code)
	if err != nil {
		log.Println("QueryGroup fail")
		return nil, nil, "", "", err
	}

	if len(retGroup) == 0 {
		return nil, nil, "", "", errors.New("invalid code")
	}

	// query players with root_id and parent_id relation
	retPlayer, err := utils.QueryPlayersByName(pl)
	if err != nil {
		log.Println("QueryPlayersByName fail", err)
		return nil, nil, "", "", err
	}

	dateBegin := date
	dateEnd, err := utils.GetNextMonthDate(date)
	if err != nil {
		log.Println("next month date invalid")
		return nil, nil, "", "", err
	}

	if half == "fh" {
		dateBegin, err = utils.GetPreMonthDate(date)
		dateEnd, err = utils.GetMidMonthDate(date)
	} else if half == "sh" {
		dateBegin, err = utils.GetMidMonthDate(date)
		dateEnd, err = utils.GetEndMonthDate(date)
	} else if half == "w" {
		dateBegin = date
		dateEnd, err = utils.GetNextWeekDate(date)
	}

	var pls []string
	for _, onePlayer := range retPlayer {
		pls = append(pls, onePlayer.Name)
	}

	// query paipu
	retPaipu, err := utils.QueryGroupPlayerPaipu(retGroup[0].Group_Id, pls, dateBegin, dateEnd)
	if err != nil {
		log.Println("QueryGroupPlayerPaipu fail")
		return nil, nil, "", "", err
	}

	return retPaipu, pls, dateBegin, dateEnd, nil
}
