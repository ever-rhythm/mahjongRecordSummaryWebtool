package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	logFile, err := os.Create("./gin.log")
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	gin.DisableConsoleColor()
	gin.DefaultWriter = io.MultiWriter(logFile)
	r := gin.Default()

	// html
	r.LoadHTMLFiles("web/index.html")
	r.GET("/index.html", func(c *gin.Context) {
		c.HTML(200, "index.html", "index")
		return
	})

	// test done
	r.POST("/api/v2", func(c *gin.Context) {
		// get input json
		b, _ := c.GetRawData()
		log.Println(string(b))

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
			records = append(records, tmpRecord)
		}

		mode, err := req.Get("mode").String()
		if err != nil {
			log.Println(err)
			c.JSON(500, "mode invalid")
			return
		}

		// validate
		uuids, err := utils.GetUuidByRecordUrl(records)
		if err != nil {
			log.Println(err)
			c.JSON(500, "url invalid")
			return
		}

		ratePt, rateZhuyi, err := utils.GetRateZhuyiByMode(mode)
		if err != nil {
			log.Println(err)
			c.JSON(500, "rate invalid")
			return
		}

		log.Println(uuids, ratePt, rateZhuyi)

		// calc
		mapPlayerInfo, ptRows, zhuyiRows, err := api.GetSummaryByUuids(uuids, ratePt, rateZhuyi)
		if err != nil {
			log.Println("GetSummaryByUuids fail ", err)
			c.JSON(500, err)
			return
		}

		// output

		// names
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
		objJs.Set("msg", "ok")
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

		c.JSON(200, tmpMap)
		return
	})

	r.POST("/api/test", func(c *gin.Context) {
		b, _ := c.GetRawData()
		fmt.Println(string(b))
		req, err := simplejson.NewJson(b)
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}

		var records []string
		mapRecords, err := req.Get("comboRecords").Array()
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}
		for _, v := range mapRecords {
			tmpMap, _ := v.(map[string]interface{})
			tmpRecord, _ := tmpMap["url"].(string)
			records = append(records, tmpRecord)
		}

		mode, err := req.Get("mode").String()
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}

		//fmt.Println(records, mode)
		mapPlayerInfo, ptRows, zhuyiRows, err := api.GetSummaryByRecords(records, mode)
		if err != nil {
			c.JSON(500, err)
		}

		// names
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
		objJs.Set("msg", "ok")
		objJs.SetPath([]string{"data", "names"}, arrName)
		objJs.SetPath([]string{"data", "rows"}, rows)
		objJs.SetPath([]string{"data", "ptRows"}, ptRows)
		objJs.SetPath([]string{"data", "zhuyiRows"}, zhuyiRows)

		tmpMap, err := objJs.Map()
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}

		//fmt.Println(tmpMap)
		c.JSON(200, tmpMap)

	})

	r.Run(":8085")
	//r.Run(":8086")
}

/*
	var rows []map[string]string

	var row = make(map[string]string)
	row["desc"] = "总计点数"
	row["p0"] = "5"
	row["p1"] = "15"
	row["p2"] = "-5"
	row["p3"] = "-15"
	rows = append(rows, row)

	row = make(map[string]string)
	row["desc"] = "总计祝仪"
	row["p0"] = "1"
	row["p1"] = "0"
	row["p2"] = "-3"
	row["p3"] = "2"
	rows = append(rows, row)

	row = make(map[string]string)
	row["desc"] = "最终积分"
	row["p0"] = "100"
	row["p1"] = "0"
	row["p2"] = "-300"
	row["p3"] = "200"
	rows = append(rows, row)

	js := simplejson.New()
	js.Set("status",0)
	js.Set("msg","ok")
	js.SetPath([]string{"data","names"}, []string{"雀士1","雀士2","雀士3","雀士4"})
	js.SetPath([]string{"data","rows"}, rows)
	js.SetPath([]string{"data","ptRows"}, rows)
	js.SetPath([]string{"data","zhuyiRows"}, rows)

	tmpMap, err := js.Map()
	if err != nil{
		fmt.Println(err)
	}

	//fmt.Println(tmpMap)
	c.JSON(200, tmpMap)

*/
