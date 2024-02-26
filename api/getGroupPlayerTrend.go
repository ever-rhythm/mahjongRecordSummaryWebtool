package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

func GetGroupPlayerTrend(code string, pl string, date string) ([]utils.TablePaipu, []string, error) {

	// query code
	retGroup, err := utils.QueryGroup(code)
	if err != nil {
		log.Println("QueryGroup fail")
		return nil, nil, err
	}

	if len(retGroup) == 0 {
		return nil, nil, errors.New("invalid code")
	}

	// query player
	retPlayer, err := utils.QueryPlayersByName(pl)
	if err != nil {
		log.Println("QueryPlayersByName fail", err)
		return nil, nil, err
	}

	dateEnd, err := utils.GetNextMonthDate(date)
	if err != nil {
		log.Println("next month date invalid")
		return nil, nil, err
	}

	var pls []string
	for _, onePlayer := range retPlayer {
		pls = append(pls, onePlayer.Name)
	}

	// query paipu
	retPaipu, err := utils.QueryGroupPlayerPaipu(retGroup[0].Group_Id, pls, date, dateEnd)
	if err != nil {
		log.Println("QueryGroupPlayerPaipu fail")
		return nil, nil, err
	}

	return retPaipu, pls, nil
}
