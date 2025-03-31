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

// v3 summary
func AdcSummary(c *gin.Context) {
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
