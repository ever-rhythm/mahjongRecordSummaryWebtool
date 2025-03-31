package action

import (
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
)

func GroupRank(c *gin.Context) {
	objJs := okRet()

	b, _ := c.GetRawData()
	req, err := simplejson.NewJson(b)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "请求格式有误")
		c.JSON(500, objJs)
		return
	}

	code := req.Get("code").MustString()
	date := req.Get("date").MustString()
	half := req.Get("half").MustString()

	if len(code) == 0 || len(date) == 0 {
		objJs.Set("msg", "请求参数有误")
		c.JSON(500, objJs)
		return
	}

	ret, dateBegin, dateEnd, err := api.GetGroupRank(code, date, half)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "查询有误")
		c.JSON(500, objJs)
		return
	}

	var players []string
	var totals []int
	var negaTotals []int

	for _, oneRank := range ret {
		players = append(players, oneRank.Pl)
		if oneRank.Total > 0 {
			totals = append(totals, oneRank.Total)
			negaTotals = append(negaTotals, 0)
		} else {
			totals = append(totals, 0)
			negaTotals = append(negaTotals, oneRank.Total)
		}
	}

	objJs.SetPath([]string{"data", "player"}, players)
	objJs.SetPath([]string{"data", "lineTotal"}, totals)
	objJs.SetPath([]string{"data", "lineNegaTotal"}, negaTotals)
	objJs.SetPath([]string{"data", "dateBegin"}, utils.GetShortDisplayDate(dateBegin))
	objJs.SetPath([]string{"data", "dateEnd"}, utils.GetShortDisplayDate(dateEnd))

	c.JSON(200, objJs.MustMap())
	return
}
