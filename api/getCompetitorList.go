package api

import (
	"github.com/mahjongRecordSummaryWebtool/utils"
	"sort"
)

type Competitor struct {
	Pl     string
	Cnt    int
	Pt     int
	Zy     int
	Total  int
	Bpt    int
	Bzy    int
	Btotal int
}

type Competitors []Competitor

func (players Competitors) Len() int {
	return len(players)
}

func (players Competitors) Swap(i, j int) {
	players[i], players[j] = players[j], players[i]
}

func (players Competitors) Less(i, j int) bool {
	//return players[j].Cnt > players[i].Cnt
	return players[j].Btotal > players[i].Btotal
}

func GetCompetitorList(pl string) ([]Competitor, error) {

	pts, err := utils.QueryPtByConds(pl, 0, 0)
	if err != nil {
		return nil, err
	}

	mapPlInfo := map[string]*Competitor{}

	for i := 0; i < len(pts); i++ {
		onePls := []string{pts[i].Pl_1, pts[i].Pl_2, pts[i].Pl_3, pts[i].Pl_4}
		for j := 0; j < len(onePls); j++ {
			_, b := mapPlInfo[onePls[j]]
			if !b && onePls[j] != pl && len(onePls[j]) > 0 {
				oneCompetitor := Competitor{Pl: onePls[j]}
				mapPlInfo[onePls[j]] = &oneCompetitor
			}
		}
	}

	//log.Println(mapPlInfo) // test

	for i := 0; i < len(pts); i++ {
		onePls := []string{pts[i].Pl_1, pts[i].Pl_2, pts[i].Pl_3, pts[i].Pl_4}
		onePts := []int{pts[i].Pt_1, pts[i].Pt_2, pts[i].Pt_3, pts[i].Pt_4}
		oneZys := []int{pts[i].Zy_1, pts[i].Zy_2, pts[i].Zy_3, pts[i].Zy_4}

		curPlIdx := 0
		curPt := 0
		curZy := 0
		for j := 0; j < len(onePls); j++ {
			if onePls[j] == pl {
				curPlIdx = j
				curPt = onePts[j]
				curZy = oneZys[j]
				break
			}
		}

		for j := 0; j < len(onePls); j++ {
			if j == curPlIdx || len(onePls[j]) == 0 {
				continue
			} else {
				mapPlInfo[onePls[j]].Pt += onePts[j]
				mapPlInfo[onePls[j]].Zy += oneZys[j]
				mapPlInfo[onePls[j]].Total += onePts[j]/100 + oneZys[j]*30
				mapPlInfo[onePls[j]].Cnt += 1
				mapPlInfo[onePls[j]].Bpt += curPt - onePts[j]
				mapPlInfo[onePls[j]].Bzy += curZy - oneZys[j]
				mapPlInfo[onePls[j]].Btotal += (curPt-onePts[j])/100 + (curZy-oneZys[j])*30
			}
		}
	}

	//log.Println(mapPlInfo) // test

	var arrCompetitor []Competitor
	for _, v := range mapPlInfo {
		arrCompetitor = append(arrCompetitor, *v)
	}

	sort.Sort(Competitors(arrCompetitor))

	//log.Println(arrCompetitor) // test

	return arrCompetitor, err
}
