package api

import (
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

// undo dev
func GetPlayerActionSummary(code string, pl string, dateStart string, dateEnd string) ([]StPtSummary, error) {

	retGroup, err := utils.QueryGroup(code)
	if err != nil || len(retGroup) == 0 {
		log.Println("query group fail")
		return nil, err
	}

	return nil, err
}
