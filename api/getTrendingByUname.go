package api

import "github.com/mahjongRecordSummaryWebtool/utils"

func GetTrendingByUname(pl string) ([]utils.Table_pt, error) {

	// select all by uname and time
	pts, err := utils.QueryPtByConds(pl, 0, 0)
	if err != nil {
		return nil, err
	}

	return pts, err
}
