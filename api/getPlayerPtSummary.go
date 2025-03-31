package api

import (
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

type StPtSummary struct {
	Total_Deal  int
	Total_Score int
	Pt          int
	Zy          int
	Zy_Yifa     int
	Zy_Li       int
	Zy_Aka      int
	Rank_1      int
	Rank_2      int
	Rank_3      int
	Rank_4      int
}

// undo dev
func GetPlayerPtSummary(code string, pl string, dateStart string, dateEnd string) ([]StPtSummary, error) {

	retGroup, err := utils.QueryGroup(code)
	if err != nil || len(retGroup) == 0 {
		log.Println("query group fail")
		return nil, err
	}

	return nil, err
}
