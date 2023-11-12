package main

import (
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

	// init majsoul server config
	err = initMajsoulConfig()
	if err != nil {
		log.Println("init majsoul server config fail", client.MajsoulServerConfig, err)
		os.Exit(0)
	}
	log.Println("init majsoul server config ok", client.MajsoulServerConfig)

	// static
	r.StaticFS("/static", http.Dir("./web/static"))

	// gin get
	r.LoadHTMLFiles("web/index.html", "web/debug.html", "web/trend.html")
	r.GET("/index.html", func(c *gin.Context) {
		c.HTML(200, "index.html", "index")
		return
	})

	r.GET("/debug.html", func(c *gin.Context) {
		c.HTML(200, "debug.html", "debug")
		return
	})

	r.GET("/trend.html", func(c *gin.Context) {
		c.HTML(200, "trend.html", "trend")
		return
	})

	// gin post
	r.POST("/api/v2", func(c *gin.Context) {
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

		//log.Println("params", uuids, ratePt, rateZhuyi)

		// calc
		mapPlayerInfo, ptRows, zhuyiRows, err := api.GetSummaryByUuids(uuids, ratePt, rateZhuyi)
		if err != nil {
			log.Println("GetSummaryByUuids err ", err)
			objJsMsg.Set("msg", err.Error())
			c.JSON(200, objJsMsg)
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

		// todo print json log
		//tmpLog, err := objJs.String()
		//log.Println("v2 ok", tmpLog)
		c.JSON(200, tmpMap)
		return
	})

	// gin post todo dev test
	r.POST("/api/trend", func(c *gin.Context) {

		tmpArr := [...]int{65, 63, 10, 73, 42, 21}

		objJs := simplejson.New()
		objJs.Set("status", 0)
		objJs.Set("msg", "成功")
		objJs.SetPath([]string{"data", "line"}, tmpArr)

		tmpMap, err := objJs.Map()
		if err != nil {
			log.Println(err)
			c.JSON(500, "js map err")
			return
		}

		c.JSON(200, tmpMap)
		return
	})

	// gin server
	err = r.Run(":8085")
	if err != nil {
		log.Println(err)
	}
}

// custom gateway
func initMajsoulConfig() error {

	client.MajsoulServerConfig.Host = "https://game.maj-soul.com"
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

	// get region
	rspRegion, err := request.Get("1/v" + client.MajsoulServerConfig.Version + "/config.json")
	if err != nil {
		log.Println("get region fail", client.MajsoulServerConfig, err)
		return err
	}

	objJsRegion, err := simplejson.NewJson(rspRegion)
	if err != nil {
		log.Println("unpack region fail", err)
		return err
	}

	client.MajsoulServerConfig.Region_urls[0], err = objJsRegion.Get("ip").GetIndex(0).Get("region_urls").GetIndex(0).Get("url").String()
	if err != nil {
		log.Println("unpack region json fail", err)
		return err
	}

	// get ws
	/*
		requestLb := utils.NewRequest(client.MajsoulServerConfig.Region_urls[0])
		rspWs, err := requestLb.Get(fmt.Sprintf("?service=ws-gateway&protocol=ws&ssl=true&rv=%d", randv))
		if err != nil {
			log.Println("get ws fail", client.MajsoulServerConfig, err)
			return err
		}

		objJsWs, err := simplejson.NewJson(rspWs)
		if err != nil {
			log.Println("unpack ws fail", err)
			return err
		}

		tmpWs, err := objJsWs.Get("servers").GetIndex(0).String()
		client.MajsoulServerConfig.Ws_server = "wss://" + tmpWs + "/gateway"
		if err != nil {
			log.Println("unpack ws json fail", err)
			return err
		}

	*/

	// custom ws todo dev test
	client.MajsoulServerConfig.Ws_server = "wss://gateway-sy.maj-soul.com:443/gateway"
	//client.MajsoulServerConfig.Ws_server = "wss://gateway-hw.maj-soul.com:443/gateway"
	return nil
}
