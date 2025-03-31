package action

import (
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

func GroupPlayerOpponentTrend(c *gin.Context) {
	objJs := okRet()

	b, _ := c.GetRawData()
	req, err := simplejson.NewJson(b)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "请求格式有误")
		c.JSON(500, objJs.MustMap())
		return
	}

	code := req.Get("code").MustString()
	date := req.Get("date").MustString()
	pl := req.Get("player").MustString()
	op := req.Get("opponent").MustString()

	if len(code) == 0 || len(pl) == 0 || len(date) == 0 || len(op) == 0 {
		objJs.Set("msg", "请求参数有误")
		c.JSON(500, objJs.MustMap())
		return
	}

	ret, retPls, err := api.GetGroupPlayerOpponentTrend(code, date, pl, op)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "查询有误")
		c.JSON(500, objJs)
		return
	}

	var idxs []int
	var pts []int
	var zys []int
	var totals []int
	cntPlacing := [4]int{}
	curPt := 0
	curZy := 0
	curTotal := 0
	var paipuDetails []utils.StTrendPaipu

	for k, onePaipu := range ret {
		oneDetail := utils.StTrendPaipu{
			Paipu_Url: onePaipu.Paipu_Url,
			Pl_1:      onePaipu.Pl_1,
			Pl_2:      onePaipu.Pl_2,
			Pl_3:      onePaipu.Pl_3,
			Pl_4:      onePaipu.Pl_4,
			Pt_1:      onePaipu.Pt_1,
			Pt_2:      onePaipu.Pt_2,
			Pt_3:      onePaipu.Pt_3,
			Pt_4:      onePaipu.Pt_4,
			Zy_1:      onePaipu.Zy_1,
			Zy_2:      onePaipu.Zy_2,
			Zy_3:      onePaipu.Zy_3,
			Zy_4:      onePaipu.Zy_4,
		}
		paipuDetails = append(paipuDetails, oneDetail)

		ratePt, rateZy, err := utils.GetRateZhuyiByMode(onePaipu.Rate)
		if err != nil {
			ratePt = 0
			rateZy = 0
		}

		arrPl := []string{onePaipu.Pl_1, onePaipu.Pl_2, onePaipu.Pl_3, onePaipu.Pl_4}
		arrPt := []int{onePaipu.Pt_1, onePaipu.Pt_2, onePaipu.Pt_3, onePaipu.Pt_4}
		arrZy := []int{onePaipu.Zy_1, onePaipu.Zy_2, onePaipu.Zy_3, onePaipu.Zy_4}

		for i := 0; i < len(arrPl); i++ {
			for j := 0; j < len(retPls); j++ {
				if arrPl[i] == retPls[j] {
					onePt := ratePt * arrPt[i] / 1000
					oneZy := rateZy * arrZy[i] * ratePt
					cntPlacing[i] += 1
					curPt += onePt
					curZy += oneZy
					curTotal += onePt + oneZy

					pts = append(pts, curPt)
					zys = append(zys, curZy)
					totals = append(totals, curTotal)
					idxs = append(idxs, k)
					break
				}
			}
		}
	}

	objJs.SetPath([]string{"data", "date"}, idxs)
	objJs.SetPath([]string{"data", "linePt"}, pts)
	objJs.SetPath([]string{"data", "lineZy"}, zys)
	objJs.SetPath([]string{"data", "lineTotal"}, totals)
	objJs.SetPath([]string{"data", "cntTotal"}, len(totals))
	objJs.SetPath([]string{"data", "cnt1"}, cntPlacing[0])
	objJs.SetPath([]string{"data", "cnt2"}, cntPlacing[1])
	objJs.SetPath([]string{"data", "cnt3"}, cntPlacing[2])
	objJs.SetPath([]string{"data", "cnt4"}, cntPlacing[3])
	objJs.SetPath([]string{"data", "lineDetail"}, paipuDetails)

	c.JSON(200, objJs.MustMap())
	return
}
