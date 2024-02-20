package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"sort"
)

func GetGroupRank(code string, date string) ([]Rank, error) {

	// query group code
	retGroup, err := utils.QueryGroup(code)
	if err != nil {
		log.Println("query group fail")
		return nil, err
	}

	if len(retGroup) == 0 {
		return nil, errors.New("invalid code")
	}

	// query player
	retPlayer, err := utils.QueryPlayer(retGroup[0].Group_Id)
	if err != nil {
		log.Println("query player fail")
		return nil, err
	}

	//log.Println("len player", len(retPlayer))

	mapPlayer2Rank := make(map[string]*Rank)
	for _, onePlayer := range retPlayer {
		mapPlayer2Rank[onePlayer.Name] = &Rank{
			Pl: onePlayer.Name,
		}
	}

	// query paipu
	retPaipu, err := utils.QueryPaipu(retGroup[0].Group_Id, date)
	if err != nil {
		log.Println("query paipu fail")
		return nil, err
	}

	//log.Println("len paipu", len(retPaipu))

	for _, onePaipu := range retPaipu {
		arrPl := []string{onePaipu.Pl_1, onePaipu.Pl_2, onePaipu.Pl_3, onePaipu.Pl_4}
		arrPt := []int{onePaipu.Pt_1, onePaipu.Pt_2, onePaipu.Pt_3, onePaipu.Pt_4}
		arrZy := []int{onePaipu.Zy_1, onePaipu.Zy_2, onePaipu.Zy_3, onePaipu.Zy_4}

		ratePt, rateZy, err := utils.GetRateZhuyiByMode(onePaipu.Rate)
		if err != nil {
			ratePt = 0
			rateZy = 0
		}

		for i := 0; i < len(arrPl); i++ {
			_, exist := mapPlayer2Rank[arrPl[i]]
			if exist {
				mapPlayer2Rank[arrPl[i]].Pt += arrPt[i] * ratePt / 1000
				mapPlayer2Rank[arrPl[i]].Zy += arrZy[i] * rateZy * ratePt
				mapPlayer2Rank[arrPl[i]].Cnt += 1
				mapPlayer2Rank[arrPl[i]].Total += ratePt*arrPt[i]/1000 + rateZy*arrZy[i]*ratePt
			}
		}
	}

	// sort rank
	var arrRank []Rank
	for _, v := range mapPlayer2Rank {
		if v.Cnt > 0 {
			arrRank = append(arrRank, *v)
		}
	}
	sort.Sort(Ranks(arrRank))

	return arrRank, nil
}
