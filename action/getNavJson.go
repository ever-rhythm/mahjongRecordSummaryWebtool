package action

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"time"
)

func GetNavJson(c *gin.Context) {
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
	oneTmp["label"] = "排行榜列表（点击下面跳转）"
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
				oneMonthSh["to"] = fmt.Sprintf("/rank.html?date=2025-%s-01&half=%s&code=%s", strMonth, oneHalfSh, code)
				arrTmp = append(arrTmp, oneMonthSh)
			}

			// first half
			oneHalf := "fh"
			oneMonth["to"] = fmt.Sprintf("/rank.html?date=2025-%s-01&half=%s&code=%s", strMonth, oneHalf, code)
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

					oneMonth["to"] = fmt.Sprintf("/rank.html?date=2025-%s-%s&half=%s&code=%s", strMonth, strDay, "w", code)
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
