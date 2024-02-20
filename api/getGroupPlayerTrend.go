package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

func GetGroupPlayerTrend(code string, pl string, date string) ([]utils.TablePaipu, error) {

	// query code
	retGroup, err := utils.QueryGroup(code)
	if err != nil {
		log.Println("QueryGroup fail")
		return nil, err
	}

	if len(retGroup) == 0 {
		return nil, errors.New("invalid code")
	}

	// query paipu
	dateEnd, err := utils.GetNextMonthDate(date)
	if err != nil {
		log.Println("next month date invalid")
		return nil, err
	}

	retPaipu, err := utils.QueryGroupPlayerPaipu(retGroup[0].Group_Id, pl, date, dateEnd)
	if err != nil {
		log.Println("QueryGroupPlayerPaipu fail")
		return nil, err
	}

	return retPaipu, nil
}
