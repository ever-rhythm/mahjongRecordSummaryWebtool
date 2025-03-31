package action

import (
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"log"
)

func Competitor(c *gin.Context) {
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

	if len(code) == 0 || len(pl) == 0 || len(date) == 0 {
		objJs.Set("msg", "请求参数有误")
		c.JSON(500, objJs.MustMap())
		return
	}

	ret, err := api.GetCompetitors(code, pl, date)
	if err != nil {
		log.Println("GetCompetitors fail", err)
		c.JSON(500, objJs.MustMap())
		return
	}

	var pls []string
	var pts []int
	var zys []int
	var pers []int
	var cnts []int
	var totals []int
	var bTotals []int
	var bPts []int
	var bZys []int

	for _, oneCompetitor := range ret {
		pls = append(pls, oneCompetitor.Pl)
		pts = append(pts, oneCompetitor.Pt)
		zys = append(zys, oneCompetitor.Zy)
		totals = append(totals, oneCompetitor.Total)
		pers = append(pers, oneCompetitor.Btotal/oneCompetitor.Cnt)
		cnts = append(cnts, oneCompetitor.Cnt)
		bPts = append(bPts, oneCompetitor.Bpt)
		bZys = append(bZys, oneCompetitor.Bzy)
		bTotals = append(bTotals, oneCompetitor.Btotal)
	}

	objJs.SetPath([]string{"data", "pls"}, pls)
	objJs.SetPath([]string{"data", "linePt"}, pts)
	objJs.SetPath([]string{"data", "lineZy"}, zys)
	objJs.SetPath([]string{"data", "linePer"}, pers)
	objJs.SetPath([]string{"data", "lineCnt"}, cnts)
	objJs.SetPath([]string{"data", "lineTotal"}, totals)
	objJs.SetPath([]string{"data", "lineBtotal"}, bTotals)
	objJs.SetPath([]string{"data", "lineBpt"}, bPts)
	objJs.SetPath([]string{"data", "lineBzy"}, bZys)

	c.JSON(200, objJs.MustMap())
	return
}
