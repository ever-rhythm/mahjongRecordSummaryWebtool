package api

import (
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

// todo test
func GetPlayerPtSummary(code string, pl string, dateStart string, dateEnd string) ([]utils.StPtSummary, error) {

	retGroup, err := utils.QueryGroup(code)
	if err != nil || len(retGroup) == 0 {
		log.Println("query group fail")
		return nil, err
	}

	rets, err := utils.QueryPtSummary(retGroup[0].Group_Id, pl, dateStart, dateEnd)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return rets, nil
}
