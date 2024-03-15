package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"sort"
)

type Rank struct {
	Pl             string
	PlayerId       int
	ParentPlayerId int
	Rate           int
	Cnt            int
	Pt             int
	Zy             int
	Total          int
	Bpt            int
	Bzy            int
	Btotal         int
}

type Ranks []Rank

func (players Ranks) Len() int {
	return len(players)
}

func (players Ranks) Swap(i, j int) {
	players[i], players[j] = players[j], players[i]
}

func (players Ranks) Less(i, j int) bool {
	return players[j].Total > players[i].Total
}

func GetGroupRank(code string, date string, half string) ([]Rank, error) {

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
			Pl:             onePlayer.Name,
			PlayerId:       onePlayer.Player_Id,
			ParentPlayerId: onePlayer.Parent_Player_Id,
		}
	}

	// query paipu
	dateEnd, err := utils.GetNextMonthDate(date)
	if err != nil {
		log.Println("get dateEnd fail", date)
		return nil, err
	}

	if half == "fh" {
		dateEnd, err = utils.GetMidMonthDate(date)
	} else if half == "sh" {
		date, err = utils.GetMidMonthDate(date)
	}

	retPaipu, err := utils.QueryPaipu(retGroup[0].Group_Id, date, dateEnd)
	if err != nil {
		log.Println("query paipu fail")
		return nil, err
	}

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
				onePt := arrPt[i] * ratePt / 1000
				oneZy := arrZy[i] * rateZy * ratePt

				if mapPlayer2Rank[arrPl[i]].ParentPlayerId != 0 {
					for _, oneRank := range mapPlayer2Rank {
						if oneRank.PlayerId == mapPlayer2Rank[arrPl[i]].ParentPlayerId {
							mapPlayer2Rank[oneRank.Pl].Pt += onePt
							mapPlayer2Rank[oneRank.Pl].Zy += oneZy
							mapPlayer2Rank[oneRank.Pl].Cnt += 1
							mapPlayer2Rank[oneRank.Pl].Total += onePt + oneZy
							break
						}
					}
				} else {
					mapPlayer2Rank[arrPl[i]].Pt += onePt
					mapPlayer2Rank[arrPl[i]].Zy += oneZy
					mapPlayer2Rank[arrPl[i]].Cnt += 1
					mapPlayer2Rank[arrPl[i]].Total += onePt + oneZy
				}
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
