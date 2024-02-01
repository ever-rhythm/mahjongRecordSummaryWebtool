package api

import (
	"github.com/mahjongRecordSummaryWebtool/utils"
	"sort"
)

type Rank struct {
	Pl     string
	Cnt    int
	Pt     int
	Zy     int
	Total  int
	Bpt    int
	Bzy    int
	Btotal int
}

type Ranks []Rank

func (players Ranks) Len() int {
	return len(players)
}

func (players Ranks) Swap(i, j int) {
	players[i], players[j] = players[j], players[i]
}

func (players Ranks) Less(i, j int) bool {
	//return players[j].Cnt > players[i].Cnt
	return players[j].Total > players[i].Total
}

func GetRankingList() ([]Rank, error) {
	// init care list
	/*
		var mapCarePl = make(map[string]int)
		var arrCarePl = []string{}

		for i:=0;i<len(arrCarePl);i++{
			mapCarePl[arrCarePl[i]] = 1
		}

	*/

	// select all by uname and time
	pts, err := utils.QueryPts(0, 0)
	if err != nil {
		return nil, err
	}

	mapPlInfo := map[string]*Rank{}

	for i := 0; i < len(pts); i++ {
		onePls := []string{pts[i].Pl_1, pts[i].Pl_2, pts[i].Pl_3, pts[i].Pl_4}
		for j := 0; j < len(onePls); j++ {
			_, b := mapPlInfo[onePls[j]]
			if !b && len(onePls[j]) > 0 {
				oneRank := Rank{Pl: onePls[j]}
				mapPlInfo[onePls[j]] = &oneRank
			}
		}
	}

	for i := 0; i < len(pts); i++ {
		onePls := []string{pts[i].Pl_1, pts[i].Pl_2, pts[i].Pl_3, pts[i].Pl_4}
		onePts := []int{pts[i].Pt_1, pts[i].Pt_2, pts[i].Pt_3, pts[i].Pt_4}
		oneZys := []int{pts[i].Zy_1, pts[i].Zy_2, pts[i].Zy_3, pts[i].Zy_4}

		for j := 0; j < len(onePls); j++ {
			if len(onePls[j]) > 0 {
				mapPlInfo[onePls[j]].Pt += onePts[j]
				mapPlInfo[onePls[j]].Zy += oneZys[j]
				mapPlInfo[onePls[j]].Total += onePts[j]/100 + oneZys[j]*30
				mapPlInfo[onePls[j]].Cnt += 1
			}
		}
	}

	var arrRank []Rank
	for _, v := range mapPlInfo {
		arrRank = append(arrRank, *v)
	}

	sort.Sort(Ranks(arrRank))

	return arrRank, err
}
