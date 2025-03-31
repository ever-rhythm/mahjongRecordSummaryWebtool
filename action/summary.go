package action

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"strconv"
)

func Summary(c *gin.Context) {
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

	// calc
	mapPlayerInfo, ptRows, zhuyiRows, err := api.GetSummaryByUuids(uuids, ratePt, rateZhuyi)
	if err != nil {
		log.Println("GetSummaryByUuids err", err)
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

	rowPtSum["desc"] = "点数"
	rowZhuyiSum["desc"] = "祝仪"
	rowMoneySum["desc"] = "总分"

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
