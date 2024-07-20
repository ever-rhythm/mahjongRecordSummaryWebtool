package api

import "github.com/mahjongRecordSummaryWebtool/utils"

type StCareer struct {
	Pl         string
	TotalDeal  int
	TotalScore int
	TotalPt    int
	TotalZy    int
	DateStart  string
	DateEnd    string
	Rank_1     int
	Rank_2     int
	Rank_3     int
	Rank_4     int
	ZyYifa     int
	ZyLi       int
	ZyAka      int
}

// todo test
func GetCareer(code string, pl string) (*StCareer, error) {

	// query vip code
	retVip, err := utils.QueryVipCode(code, pl)
	if err != nil || len(retVip) == 0 {
		return nil, err
	}

	// get total
	retSummary, err := utils.QueryTotalSummary(pl)
	if err != nil || len(retSummary) == 0 {
		return nil, err
	}

	retCareer := &StCareer{
		Pl:         pl,
		TotalDeal:  retSummary[0].Total_Deal,
		TotalScore: retSummary[0].Total_Score,
		TotalPt:    retSummary[0].Pt,
		TotalZy:    retSummary[0].Zy,
		Rank_1:     retSummary[0].Rank_1,
		Rank_2:     retSummary[0].Rank_2,
		Rank_3:     retSummary[0].Rank_3,
		Rank_4:     retSummary[0].Rank_4,
	}

	return retCareer, nil
}
