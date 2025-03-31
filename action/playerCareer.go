package action

import (
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"time"
)

func PlayerCareer(c *gin.Context) {
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
	pl := req.Get("player").MustString()

	// default current month
	timeNow := time.Now()
	timeIndex := time.Date(timeNow.Year(), timeNow.Month(), 1, 0, 0, 0, 0, time.Local)
	timeSpan := utils.TIME_SPAN_TOTAL

	retCareer, err := api.GetCareer(code, pl, timeIndex, timeSpan)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "请求有误")
		c.JSON(500, objJs.MustMap())
		return
	}

	dateCurMonth := utils.GetCurMonthDate()
	dateNextMonth, _ := utils.GetNextMonthDate(dateCurMonth)

	objJs.SetPath([]string{"data", "pl"}, pl)
	objJs.SetPath([]string{"data", "dateBegin"}, dateCurMonth)
	objJs.SetPath([]string{"data", "dateEnd"}, dateNextMonth)
	objJs.SetPath([]string{"data", "totalScore"}, retCareer.TotalScore)
	objJs.SetPath([]string{"data", "totalDeal"}, retCareer.TotalDeal)
	objJs.SetPath([]string{"data", "totalPt"}, retCareer.TotalPt)
	objJs.SetPath([]string{"data", "totalZy"}, retCareer.TotalZy)
	objJs.SetPath([]string{"data", "rank_1"}, retCareer.Rank_1)
	objJs.SetPath([]string{"data", "rank_2"}, retCareer.Rank_2)
	objJs.SetPath([]string{"data", "rank_3"}, retCareer.Rank_3)
	objJs.SetPath([]string{"data", "rank_4"}, retCareer.Rank_4)
	c.JSON(200, objJs)
	return
}
