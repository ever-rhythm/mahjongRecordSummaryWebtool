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

	// load config
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
	r.LoadHTMLFiles("web/index.html")
	r.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", "home"); return })
	r.GET("/index.html", func(c *gin.Context) { c.HTML(200, "index.html", "index"); return })

	// post
	r.POST("/api/v2", summary)

	/* bak for non-display
	// get
	r.LoadHTMLFiles("web/index.html", "web/debug.html", "web/trend.html", "web/add.html", "web/vs.html", "web/rank.html")
	r.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", "home"); return })
	r.GET("/index.html", func(c *gin.Context) { c.HTML(200, "index.html", "index"); return })
	r.GET("/debug.html", func(c *gin.Context) { c.HTML(200, "debug.html", "debug"); return })
	r.GET("/trend.html", func(c *gin.Context) { c.HTML(200, "trend.html", "trend"); return })
	r.GET("/add.html", func(c *gin.Context) { c.HTML(200, "add.html", "add"); return })
	r.GET("/vs.html", func(c *gin.Context) { c.HTML(200, "vs.html", "vs"); return })
	r.GET("/rank.html", func(c *gin.Context) { c.HTML(200, "rank.html", "rank"); return })

	// post
	r.POST("/api/v2", summary)
	r.POST("/api/addMajsoulRecord", addMajsoulRecord)
	r.POST("/api/trend", trend)
	r.POST("/api/competitor", competitor)
	r.POST("/api/rank", rank)

	*/

	// server
	err = r.Run(ConfigGin.Domain + ":" + ConfigGin.Port)
	if err != nil {
		log.Fatalln(err)
	}
}

func rank(c *gin.Context) {
	objJs := okRet()

	rets, err := api.GetRankingList()
	if err != nil {
		log.Println(err)
		c.JSON(500, objJs)
		return
	}

	var pts []int
	var zys []int
	var totals []int
	var dates []string
	var cnts []int
	var pers []int
	numLimit := 0

	for i := 0; i < len(rets); i++ {
		if rets[i].Cnt >= numLimit {
			dates = append(dates, rets[i].Pl)
			pts = append(pts, rets[i].Pt/100)
			zys = append(zys, rets[i].Zy*30)
			totals = append(totals, rets[i].Total)
			cnts = append(cnts, rets[i].Cnt)
			pers = append(pers, rets[i].Total/rets[i].Cnt)
		}
	}

	objJs.SetPath([]string{"data", "date"}, dates)
	objJs.SetPath([]string{"data", "linePt"}, pts)
	objJs.SetPath([]string{"data", "lineZy"}, zys)
	objJs.SetPath([]string{"data", "lineTotal"}, totals)
	objJs.SetPath([]string{"data", "lineCnt"}, cnts)
	objJs.SetPath([]string{"data", "linePer"}, pers)

	c.JSON(200, objJs.MustMap())
	return
}

func competitor(c *gin.Context) {
	objJs := okRet()

	b, _ := c.GetRawData()
	log.Println(string(b))
	req, err := simplejson.NewJson(b)
	oneName := req.Get("pl").MustString()
	//ratePt := req.Get("ratePt").MustInt()
	//rateZy := req.Get("rateZy").MustInt()

	rets, err := api.GetCompetitorList(oneName)
	if err != nil {
		log.Println(err)
		c.JSON(500, objJs)
		return
	}

	//log.Println(len(rets))
	var pts []int
	var zys []int
	var totals []int
	var dates []string
	var cnts []int
	var pers []int

	for i := 0; i < len(rets); i++ {
		dates = append(dates, rets[i].Pl)
		pts = append(pts, rets[i].Bpt/100)
		zys = append(zys, rets[i].Bzy*30)
		totals = append(totals, rets[i].Btotal)
		cnts = append(cnts, rets[i].Cnt)
		pers = append(pers, rets[i].Btotal/rets[i].Cnt)
	}

	objJs.SetPath([]string{"data", "date"}, dates)
	objJs.SetPath([]string{"data", "linePt"}, pts)
	objJs.SetPath([]string{"data", "lineZy"}, zys)
	objJs.SetPath([]string{"data", "lineTotal"}, totals)
	objJs.SetPath([]string{"data", "lineCnt"}, cnts)
	objJs.SetPath([]string{"data", "linePer"}, pers)

	c.JSON(200, objJs.MustMap())
	return
}

func trend(c *gin.Context) {
	objJs := okRet()
	var pts []int
	var zys []int
	var totals []int
	var dates []int
	curPt := 0
	curZy := 0
	curTotal := 0
	cnt1 := 0
	cnt2 := 0
	cnt3 := 0
	cnt4 := 0

	b, _ := c.GetRawData()
	log.Println(string(b))
	req, err := simplejson.NewJson(b)
	oneName := req.Get("pl").MustString()
	ratePt := req.Get("ratePt").MustInt()
	rateZy := req.Get("rateZy").MustInt()

	rets, err := api.GetTrendingByUname(oneName)
	if err != nil {
		log.Println(err)
		c.JSON(500, objJs)
		return
	}

	for i := 0; i < len(rets); i++ {
		var onePt, oneZy, oneRate int
		oneRate = 1
		if rets[i].Rate != 0 {
			//continue
			oneRate = rets[i].Rate / 10
		}
		if rets[i].Pl_1 == oneName {
			onePt = rets[i].Pt_1 / ratePt * oneRate
			oneZy = rets[i].Zy_1 * rateZy * oneRate
			cnt1 += 1
		} else if rets[i].Pl_2 == oneName {
			onePt = rets[i].Pt_2 / ratePt * oneRate
			oneZy = rets[i].Zy_2 * rateZy * oneRate
			cnt2 += 1
		} else if rets[i].Pl_3 == oneName {
			onePt = rets[i].Pt_3 / ratePt * oneRate
			oneZy = rets[i].Zy_3 * rateZy * oneRate
			cnt3 += 1
		} else if rets[i].Pl_4 == oneName {
			onePt = rets[i].Pt_4 / ratePt * oneRate
			oneZy = rets[i].Zy_4 * rateZy * oneRate
			cnt4 += 1
		}

		curPt += onePt
		curZy += oneZy
		curTotal += onePt + oneZy
		pts = append(pts, curPt)
		zys = append(zys, curZy)
		totals = append(totals, curTotal)
		dates = append(dates, i+1)
	}

	objJs.SetPath([]string{"data", "date"}, dates)
	objJs.SetPath([]string{"data", "linePt"}, pts)
	objJs.SetPath([]string{"data", "lineZy"}, zys)
	objJs.SetPath([]string{"data", "lineTotal"}, totals)
	objJs.SetPath([]string{"data", "cntTotal"}, len(totals))
	objJs.SetPath([]string{"data", "cnt1"}, cnt1)
	objJs.SetPath([]string{"data", "cnt2"}, cnt2)
	objJs.SetPath([]string{"data", "cnt3"}, cnt3)
	objJs.SetPath([]string{"data", "cnt4"}, cnt4)

	c.JSON(200, objJs.MustMap())
	return
}

func addMajsoulRecord(c *gin.Context) {
	objJs := simplejson.New()
	objJs.Set("status", 0)
	objJs.Set("msg", "成功")
	tmpRspMap := objJs.MustMap()

	b, _ := c.GetRawData()
	log.Println("raw data", string(b))

	req, err := simplejson.NewJson(b)
	if err != nil {
		log.Println(err)
		c.JSON(500, tmpRspMap)
		return
	}

	var records []string
	mapRecords, err := req.Get("comboRecords").Array()
	if err != nil {
		log.Println(err)
		c.JSON(500, tmpRspMap)
		return
	}
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
		c.JSON(500, tmpRspMap)
		return
	}

	err = api.AddRecordByUuid(uuids[0])
	if err != nil {
		log.Println(err)
		c.JSON(500, tmpRspMap)
		return
	}

	c.JSON(200, tmpRspMap)
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

	// output
	// build names
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
		rowZhuyiSum[oneIdx] = strconv.Itoa(int(oneInfo.Zhuyi))
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
