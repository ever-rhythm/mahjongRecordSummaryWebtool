package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"time"
)

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

func GetCareer(code string, pl string, timeIndex time.Time, timeSpan string) (*StCareer, error) {

	// query vip code
	retVip, err := utils.QueryVipCode(code, pl)
	if err != nil {
		log.Println(err)
		return nil, errors.New("查询失败")
	}
	if len(retVip) == 0 {
		return nil, errors.New("用户名和邀请码不匹配")
	}

	// query plcr
	retPlcr, err := utils.QueryPlayerCareer(pl, timeIndex, timeSpan)
	if err != nil {
		log.Println(err)
		return nil, errors.New("查询失败")
	}
	if len(retPlcr) == 0 {
		return nil, errors.New("用户名战绩不存在")
	}

	retCareer := &StCareer{
		Pl:         pl,
		TotalDeal:  retPlcr[0].Deal_Total,
		TotalScore: retPlcr[0].Score_Total,
		TotalPt:    retPlcr[0].Pt_Total,
		TotalZy:    retPlcr[0].Zy_Total,
		Rank_1:     retPlcr[0].Rank_1,
		Rank_2:     retPlcr[0].Rank_2,
		Rank_3:     retPlcr[0].Rank_3,
		Rank_4:     retPlcr[0].Rank_4,
	}

	return retCareer, nil
}
