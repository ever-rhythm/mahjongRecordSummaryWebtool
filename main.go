package main

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type configGin struct {
	Domain string
	Port   string
}

var ConfigGin = configGin{}

func main() {
	// init log
	logFile, err := os.Create("./gin.log." + time.Now().Format("2006-01-02.15-04-05"))
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(logFile)
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// load json config
	err = loadConfigFile()
	if err != nil {
		log.Fatalln("load config.json fail", err)
	}

	// init majsoul server config
	err = initMajsoulConfig()
	if err != nil {
		log.Fatalln("http majsoul server config fail", client.MajsoulServerConfig, err)
	}

	// static
	r.StaticFS("/static", http.Dir("./web/static"))

	// get
	r.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", "home"); return })
	r.GET("/index.html", func(c *gin.Context) { c.HTML(200, "index.html", "index"); return })
	r.GET("/rank.html", func(c *gin.Context) { c.HTML(200, "rank.html", "rank"); return })
	r.GET("/trend.html", func(c *gin.Context) { c.HTML(200, "trend.html", "trend"); return })
	r.GET("/debug.html", func(c *gin.Context) { c.HTML(200, "debug.html", "debug"); return })
	r.GET("/vs.html", func(c *gin.Context) { c.HTML(200, "vs.html", "vs"); return })
	r.GET("/about.html", func(c *gin.Context) { c.HTML(200, "about.html", "about"); return })
	r.GET("/nav.html", func(c *gin.Context) { c.HTML(200, "nav.html", "nav"); return })
	r.LoadHTMLFiles("web/index.html", "web/rank.html", "web/trend.html", "web/debug.html", "web/vs.html", "web/about.html", "web/nav.html")

	// post
	r.POST("/api/v2", summary)
	r.POST("/api/group_rank", groupRank)
	r.POST("/api/group_player_trend", groupPlayerTrend)
	r.POST("/api/competitor", competitor)
	r.POST("/api/group_player_op_trend", groupPlayerOpponentTrend)
	r.POST("/api/nav", getNavJson)

	r.POST("/api/plcr", playerCareer)
	r.POST("/api/ads_summary", adsSummary)

	// server
	err = r.Run(ConfigGin.Domain + ":" + ConfigGin.Port)
	if err != nil {
		log.Fatalln(err)
	}
}

func groupPlayerTrend(c *gin.Context) {
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
	pl := req.Get("player").MustString()
	half := req.Get("half").MustString()

	if len(code) == 0 || len(pl) == 0 || len(date) == 0 {
		objJs.Set("msg", "请求参数有误")
		c.JSON(500, objJs)
		return
	}

	ret, retPls, dateBegin, dateEnd, err := api.GetGroupPlayerTrend(code, pl, date, half)
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
	objJs.SetPath([]string{"data", "dateBegin"}, dateBegin)
	objJs.SetPath([]string{"data", "dateEnd"}, dateEnd)

	c.JSON(200, objJs.MustMap())
	return
}

func groupRank(c *gin.Context) {
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
	objJs.SetPath([]string{"data", "dateBegin"}, dateBegin)
	objJs.SetPath([]string{"data", "dateEnd"}, dateEnd)

	c.JSON(200, objJs.MustMap())
	return
}

// v3 summary
func adsSummary(c *gin.Context) {
	objJs := okRet()
	b, _ := c.GetRawData()
	req, err := simplejson.NewJson(b)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "请求格式有误")
		c.JSON(500, objJs.MustMap())
		return
	}

	mode, err := req.Get("mode").String()
	var records []string
	mapRecords, err := req.Get("comboRecords").Array()
	for _, v := range mapRecords {
		tmpMap, _ := v.(map[string]interface{})
		tmpRecord, _ := tmpMap["url"].(string)
		if len(tmpRecord) > 0 {
			records = append(records, tmpRecord)
		}
	}

	uuids, err := utils.GetUuidByRecordUrl(records)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "paipu url 请求格式有误")
		c.JSON(500, objJs.MustMap())
		return
	}

	// ping login and renew version code
	acc, pwd := utils.GetMajSoulBotByIdx(len(utils.ConfigMajsoulBot.Acc) - 1)
	err = api.PingMajsoulLogin(acc, pwd)
	if err != nil {
		log.Println("ping fail, renew version code", err)
		objJs.Set("msg", "连接服务器错误")
		c.JSON(500, objJs)

		err = initMajsoulConfig()
		if err != nil {
			log.Fatalln("http majsoul server config fail", client.MajsoulServerConfig, err)
		}
		return
	}

	// api get summary
	retAds, err := api.GetAdsSummary(uuids, mode)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", err)
		c.JSON(500, objJs.MustMap())
		return
	}

	// build total rows
	retTotal := api.StCalc{
		PlayerCnt: retAds[0].PlayerCnt,
		Pls:       retAds[0].Pls,
	}

	for i := 0; i < len(retAds); i++ {
		for j := 0; j < retAds[i].PlayerCnt; j++ {
			for k := 0; k < retTotal.PlayerCnt; k++ {
				if retAds[i].Pls[j] == retTotal.Pls[k] {
					retTotal.Pts[k] += retAds[i].Pts[j]
					retTotal.Zys[k] += retAds[i].Zys[j]
					retTotal.Totals[k] += retAds[i].Totals[j]
					break
				}
			}
		}
	}

	var rows []map[string]string
	var rowPtSum = make(map[string]string)
	var rowZhuyiSum = make(map[string]string)
	var rowMoneySum = make(map[string]string)

	rowPtSum["desc"] = "总计点数"
	rowZhuyiSum["desc"] = "总计祝仪"
	rowMoneySum["desc"] = "最终积分"

	for i := 0; i < retTotal.PlayerCnt; i++ {
		oneIdx := "p" + strconv.Itoa(i)
		rowPtSum[oneIdx] = fmt.Sprintf("%.1f", float32(retTotal.Pts[i])/1000)
		rowZhuyiSum[oneIdx] = strconv.Itoa(retTotal.Zys[i])
		rowMoneySum[oneIdx] = fmt.Sprintf("%.2f", retTotal.Totals[i])
		if retTotal.Totals[i] > 0 {
			rowMoneySum[oneIdx] = "+" + rowMoneySum[oneIdx]
		}
	}

	rows = append(rows, rowPtSum, rowZhuyiSum, rowMoneySum)

	// build pt rows
	var ptRows []map[string]string
	for i := 0; i < len(retAds); i++ {
		var onePtRow = make(map[string]string)
		onePtRow["desc"] = fmt.Sprintf("第%d局", i+1)

		for j := 0; j < retTotal.PlayerCnt; j++ {
			for k := 0; k < retAds[i].PlayerCnt; k++ {
				if retTotal.Pls[j] == retAds[i].Pls[k] {
					oneIdx := "p" + strconv.Itoa(j)
					onePtRow[oneIdx] = fmt.Sprintf("%.1f", float32(retAds[i].Pts[k])/1000)
					break
				}
			}
		}

		ptRows = append(ptRows, onePtRow)
	}

	// build zy rows
	var zyRows []map[string]string
	idxZyDesc := 0
	for i := 0; i < len(retAds); i++ {
		for l := 0; l < len(retAds[i].ZyDetailCnt); l++ {
			var oneZyRow = make(map[string]string)

			if i == idxZyDesc {
				oneZyRow["desc"] += fmt.Sprintf("第%d局", i+1)
				idxZyDesc += 1
			}
			oneZyRow["desc"] += retAds[i].ZyDescs[l]

			for j := 0; j < retTotal.PlayerCnt; j++ {
				for k := 0; k < retAds[i].PlayerCnt; k++ {
					if retTotal.Pls[j] == retAds[i].Pls[k] {
						oneIdx := "p" + strconv.Itoa(j)

						if retAds[i].ZyDetailCnt[l][k] != 0 {
							oneZyRow[oneIdx] = strconv.Itoa(retAds[i].ZyDetailCnt[l][k])
							if retAds[i].ZyDetailCnt[l][k] > 0 {
								oneZyRow[oneIdx] = "+" + oneZyRow[oneIdx]
							}
						} else {
							oneZyRow[oneIdx] = " "
						}
						break
					}
				}
			}

			zyRows = append(zyRows, oneZyRow)
		}
	}

	objJs.SetPath([]string{"data", "names"}, retTotal.Pls)
	objJs.SetPath([]string{"data", "rows"}, rows)
	objJs.SetPath([]string{"data", "ptRows"}, ptRows)
	objJs.SetPath([]string{"data", "zhuyiRows"}, zyRows)

	tmpLog, err := objJs.Encode()
	log.Println("api adsSummary ok ", uuids, mode, string(tmpLog))

	c.JSON(200, objJs.MustMap())
	return
}

// todo dev
func playerCareer(c *gin.Context) {
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

	retCareer, err := api.GetCareer(code, pl)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "请求有误")
		c.JSON(500, objJs.MustMap())
		return
	}

	objJs.SetPath([]string{"data", "pl"}, pl)
	objJs.SetPath([]string{"data", "totalScore"}, retCareer.TotalScore)
	objJs.SetPath([]string{"data", "TotalDeal"}, retCareer.TotalDeal)
	objJs.SetPath([]string{"data", "totalPt"}, retCareer.TotalPt)
	objJs.SetPath([]string{"data", "totalZy"}, retCareer.TotalZy)
	c.JSON(200, objJs)
	return
}

func summary(c *gin.Context) {
	// get input json
	b, _ := c.GetRawData()
	log.Println("raw data", string(b))

	objJsMsg := simplejson.New()
	objJsMsg.Set("status", 0)
	objJsMsg.Set("msg", "")

	req, err := simplejson.NewJson(b)
	if err != nil {
		log.Println(err)
		c.JSON(500, "req invalid")
		return
	}

	var records []string
	mapRecords, err := req.Get("comboRecords").Array()
	if err != nil {
		log.Println(err)
		c.JSON(500, "record invalid")
		return
	}
	for _, v := range mapRecords {
		tmpMap, _ := v.(map[string]interface{})
		tmpRecord, _ := tmpMap["url"].(string)
		if len(tmpRecord) > 0 {
			records = append(records, tmpRecord)
		}
	}

	mode, err := req.Get("mode").String()
	if err != nil {
		log.Println(err)
		objJsMsg.Set("msg", "祝仪格式有误")
		c.JSON(500, objJsMsg)
		return
	}

	// validate
	uuids, err := utils.GetUuidByRecordUrl(records)
	if err != nil {
		log.Println(err)
		objJsMsg.Set("msg", "牌谱链接有误")
		c.JSON(500, objJsMsg)
		return
	}

	ratePt, rateZhuyi, err := utils.GetRateZhuyiByMode(mode)
	if err != nil {
		log.Println(err)
		c.JSON(500, "rate invalid")
		return
	}

	// ping login , if fail then renew version code
	acc, pwd := utils.GetMajSoulBotByIdx(len(utils.ConfigMajsoulBot.Acc) - 1)
	err = api.PingMajsoulLogin(acc, pwd)
	if err != nil {
		log.Println("ping fail, renew version code", err)

		err = initMajsoulConfig()
		if err != nil {
			log.Fatalln("http majsoul server config fail", client.MajsoulServerConfig, err)
		}

		objJsMsg.Set("msg", "内部错误")
		c.JSON(500, objJsMsg)
		return
	}

	// calc
	mapPlayerInfo, ptRows, zhuyiRows, err := api.GetSummaryByUuids(uuids, ratePt, rateZhuyi)
	if err != nil {
		log.Println("GetSummaryByUuids fail", err)
		objJsMsg.Set("msg", err.Error())
		c.JSON(500, objJsMsg)
		return
	}

	// output build names
	var arrName []string
	for i := 0; i < 4; i++ {
		for _, oneInfo := range mapPlayerInfo {
			if i == int(oneInfo.Seat) {
				arrName = append(arrName, oneInfo.Nickname)
				break
			}
		}
	}

	var rows []map[string]string
	var rowPtSum = make(map[string]string)
	var rowZhuyiSum = make(map[string]string)
	var rowMoneySum = make(map[string]string)

	rowPtSum["desc"] = "总计点数"
	rowZhuyiSum["desc"] = "总计祝仪"
	rowMoneySum["desc"] = "最终积分"

	for _, oneInfo := range mapPlayerInfo {
		oneIdx := "p" + strconv.Itoa(int(oneInfo.Seat))
		rowPtSum[oneIdx] = fmt.Sprintf("%.1f", float32(oneInfo.TotalPoint)/1000)
		rowZhuyiSum[oneIdx] = strconv.Itoa(oneInfo.Zhuyi)
		rowMoneySum[oneIdx] = oneInfo.Sum
	}

	rows = append(rows, rowPtSum, rowZhuyiSum, rowMoneySum)

	objJs := simplejson.New()
	objJs.Set("status", 0)
	objJs.Set("msg", "成功")
	objJs.SetPath([]string{"data", "names"}, arrName)
	objJs.SetPath([]string{"data", "rows"}, rows)
	objJs.SetPath([]string{"data", "ptRows"}, ptRows)
	objJs.SetPath([]string{"data", "zhuyiRows"}, zhuyiRows)

	tmpMap, err := objJs.Map()
	if err != nil {
		log.Println(err)
		c.JSON(500, "js map err")
		return
	}

	tmpLog, err := objJs.Encode()
	log.Println("api v2 ok", string(tmpLog))
	c.JSON(200, tmpMap)
	return
}

func groupPlayerOpponentTrend(c *gin.Context) {
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

func competitor(c *gin.Context) {
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

func getNavJson(c *gin.Context) {
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
	arrTmp := make([]map[string]string, 0)
	oneTmp := make(map[string]string)
	oneTmp["label"] = "排行榜列表（点击下面文字跳转）"
	oneTmp["to"] = "/index.html"
	arrTmp = append(arrTmp, oneTmp)

	if code == "503zbl" {
		curMonth, _ := strconv.Atoi(time.Now().Format("01"))
		curDay, _ := strconv.Atoi(time.Now().Format("02"))

		for i := curMonth; i >= 2; i-- {
			oneMonth := make(map[string]string)
			oneMonth["label"] = fmt.Sprintf("%d月上半赛季", i)
			strMonth := strconv.Itoa(i)
			if len(strMonth) < 2 {
				strMonth = "0" + strMonth
			}

			// second half
			if i < curMonth || (i == curMonth && curDay >= 15) {
				oneMonthSh := make(map[string]string)
				oneMonthSh["label"] = fmt.Sprintf("%d月下半赛季", i)
				oneHalfSh := "sh"
				oneMonthSh["to"] = fmt.Sprintf("/rank.html?date=2024-%s-01&half=%s&code=%s", strMonth, oneHalfSh, code)
				arrTmp = append(arrTmp, oneMonthSh)
			}

			// first half
			oneHalf := "fh"
			oneMonth["to"] = fmt.Sprintf("/rank.html?date=2024-%s-01&half=%s&code=%s", strMonth, oneHalf, code)
			arrTmp = append(arrTmp, oneMonth)
		}
	} else if code == "203zbl" {
		curTime := time.Now()
		curMonth, _ := strconv.Atoi(curTime.Format("01"))
		curDay := curTime.Day()

		for i := curMonth; i >= 9; i-- {
			for j := 31; j >= 1; j-- {
				oneDay := time.Date(curTime.Year(), time.Month(i), j, 0, 0, 0, 0, time.Local)
				if (j <= curDay && curMonth == i && int(oneDay.Weekday()) == 1) || (i < curMonth && int(oneDay.Weekday()) == 1) {

					oneMonth := make(map[string]string)
					oneMonth["label"] = fmt.Sprintf("%d月%d日赛季", i, j)
					strMonth := strconv.Itoa(i)
					if len(strMonth) < 2 {
						strMonth = "0" + strMonth
					}
					strDay := strconv.Itoa(j)
					if len(strDay) < 2 {
						strDay = "0" + strDay
					}

					oneMonth["to"] = fmt.Sprintf("/rank.html?date=2024-%s-%s&half=%s&code=%s", strMonth, strDay, "w", code)
					arrTmp = append(arrTmp, oneMonth)
				}
			}
		}
	}

	if len(arrTmp) > 2 {
		arrTmp[1]["label"] += "（当前赛季）"
	}

	objJs.Set("data", arrTmp)
	c.JSON(200, objJs.MustMap())
	return
}

// load config json file
func loadConfigFile() error {

	btConfig, err := os.ReadFile("./config.json")
	if err != nil {
		log.Println("load config.json fail")
		return err
	}
	objJsConfig, err := simplejson.NewJson(btConfig)
	if err != nil {
		log.Println("conv config.json to json fail")
		return err
	}

	// set config
	ConfigGin.Domain = objJsConfig.Get("server").Get("domain").MustString()
	ConfigGin.Port = objJsConfig.Get("server").Get("port").MustString()
	client.MajsoulServerConfig.Host = objJsConfig.Get("MajsoulServerConfig").Get("host").MustString()
	client.MajsoulServerConfig.Ws_server = objJsConfig.Get("MajsoulServerConfig").Get("ws_server").MustString()
	utils.ConfigDb.Dsn = objJsConfig.Get("postgre").Get("dsn").MustString()
	utils.ConfigMajsoulBot.Acc = objJsConfig.Get("MajsoulBot").Get("acc").MustStringArray()
	utils.ConfigMajsoulBot.Pwd = objJsConfig.Get("MajsoulBot").Get("pwd").MustStringArray()
	utils.ConfigMode.Mode = objJsConfig.Get("modeMap").Get("mode").MustStringArray()
	utils.ConfigMode.Rate = objJsConfig.Get("modeMap").Get("rate").MustStringArray()
	utils.ConfigMode.Zy = objJsConfig.Get("modeMap").Get("zy").MustStringArray()
	utils.ConfigMode.RecordContestIds = objJsConfig.Get("recordContestUid").MustStringArray()
	utils.ConfigMode.RecordMode = objJsConfig.Get("recordMode").MustInt()
	utils.ConfigMode.RecordModeList = objJsConfig.Get("recordModeList").MustStringArray()
	cfgInterval := objJsConfig.Get("MajsoulServerConfig").Get("interval").MustInt()
	if cfgInterval > client.MajsoulServerConfig.Interval {
		client.MajsoulServerConfig.Interval = cfgInterval
	}

	// validate config
	if len(utils.ConfigMajsoulBot.Acc) != len(utils.ConfigMajsoulBot.Pwd) ||
		len(utils.ConfigMajsoulBot.Acc) == 0 {
		log.Println("config err, bot acc pwd len not match")
		return errors.New("bot err")
	}

	if len(utils.ConfigMode.Mode) != len(utils.ConfigMode.Rate) ||
		len(utils.ConfigMode.Mode) != len(utils.ConfigMode.Zy) ||
		len(utils.ConfigMode.Mode) == 0 {
		log.Println("config err, mode rate zy len not match")
		return errors.New("mode err")
	}

	log.Println("load config.json ok", ConfigGin)

	return nil
}

func initMajsoulConfig() error {

	// http client
	request := utils.NewRequest(client.MajsoulServerConfig.Host)
	randv := int(rand.Float32()*1000000000) + int(rand.Float32()*1000000000)

	// get version
	rspVersion, err := request.Get(fmt.Sprintf("1/version.json?randv=%d", randv))
	if err != nil {
		log.Println("get version host fail", client.MajsoulServerConfig, err)
		return err
	}

	objJsVersion, err := simplejson.NewJson(rspVersion)
	if err != nil {
		log.Println("unpack version fail", err)
		return err
	}

	client.MajsoulServerConfig.Version, err = objJsVersion.Get("version").String()
	client.MajsoulServerConfig.Force_version, err = objJsVersion.Get("force_version").String()
	client.MajsoulServerConfig.Code, err = objJsVersion.Get("code").String()

	log.Println("init majsoul server config ok", client.MajsoulServerConfig)

	return nil
}

func okRet() *simplejson.Json {
	objSucc := simplejson.New()
	objSucc.Set("status", 0)
	objSucc.Set("msg", "成功")
	return objSucc
}
