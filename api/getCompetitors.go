package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"sort"
	"strconv"
)

type Competitor struct {
	Pl     string
	Rate   int
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
	return players[j].Btotal > players[i].Btotal
}

func GetCompetitors(code string, pl string, date string) ([]Competitor, error) {

	// query code
	retGroup, err := utils.QueryGroup(code)
	if err != nil {
		log.Println("QueryGroup fail")
		return nil, err
	}

	if len(retGroup) == 0 {
		return nil, errors.New("invalid code")
	}

	dateEnd, err := utils.GetNextMonthDate(date)
	if err != nil {
		log.Println("next month date invalid")
		return nil, err
	}

	// query paipu
	pts, err := utils.QueryGroupPlayerPaipu(retGroup[0].Group_Id, []string{pl}, date, dateEnd)
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

	//log.Println(retGroup[0].Group_Id, []string{pl}, date, dateEnd, len(pts)) // test

	for i := 0; i < len(pts); i++ {
		onePls := []string{pts[i].Pl_1, pts[i].Pl_2, pts[i].Pl_3, pts[i].Pl_4}
		onePts := []int{pts[i].Pt_1, pts[i].Pt_2, pts[i].Pt_3, pts[i].Pt_4}
		oneZys := []int{pts[i].Zy_1, pts[i].Zy_2, pts[i].Zy_3, pts[i].Zy_4}
		oneRate, err := strconv.Atoi(pts[i].Rate)
		if err != nil {
			log.Println("rate err", err)
			continue
		}

		oneRatePt, oneRateZy, err := utils.GetRateZhuyiByMode(pts[i].Rate)
		if err != nil {
			log.Println("rate err", err)
			continue
		}

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
				onePt := onePts[j] * oneRatePt / 1000
				oneZy := oneZys[j] * oneRateZy * oneRatePt
				oneBpt := (curPt - onePts[j]) * oneRatePt / 1000
				oneBzy := (curZy - oneZys[j]) * oneRateZy * oneRatePt

				mapPlInfo[onePls[j]].Pt += onePt
				mapPlInfo[onePls[j]].Zy += oneZy
				mapPlInfo[onePls[j]].Total += onePt + oneZy
				mapPlInfo[onePls[j]].Cnt += 1
				mapPlInfo[onePls[j]].Rate = oneRate
				mapPlInfo[onePls[j]].Bpt += oneBpt
				mapPlInfo[onePls[j]].Bzy += oneBzy
				mapPlInfo[onePls[j]].Btotal += oneBpt + oneBzy
			}
		}
	}

	//log.Println(mapPlInfo) // test

	var arrCompetitor []Competitor
	for _, v := range mapPlInfo {
		arrCompetitor = append(arrCompetitor, *v)
	}

	sort.Sort(Competitors(arrCompetitor))

	return arrCompetitor, err
}
