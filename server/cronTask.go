package server

import (
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"strconv"
	"time"
)

// CronUpdateMajsoulConfig cron update from http to config , http renew majsoul version code
func CronUpdateMajsoulConfig() {
	for {
		time.Sleep(time.Second * time.Duration(client.MajsoulServerConfig.CodeUpdateInterval))
		curTime := time.Now().Unix()

		if curTime-client.MajsoulServerConfig.CodeUpdateTimestamp > client.MajsoulServerConfig.CodeUpdateInterval {
			client.MajsoulServerConfig.CodeUpdateTimestamp = curTime

			err := InitMajsoulConfig()
			if err != nil {
				log.Println("InitMajsoulConfig fail", client.MajsoulServerConfig, err)
				continue
			}
		}
	}
}

// todo test
func CronUpdatePlayerCareer() {
	for {
		time.Sleep(time.Second * time.Duration(ConfigCron.IntervalCronUpdatePlayerCareer))
		timeNow := time.Now()
		timeIndex := time.Date(timeNow.Year(), timeNow.Month(), 1, 0, 0, 0, 0, time.Local)
		curTime := timeNow.Unix()

		if curTime-ConfigCron.TimestampCronUpdatePlayerCareer > ConfigCron.IntervalCronUpdatePlayerCareer {
			ConfigCron.TimestampCronUpdatePlayerCareer = curTime

			retVips, err := utils.QueryVips()
			if err != nil {
				log.Println("QueryVips fail", err)
				continue
			}

			for i := 0; i < len(retVips); i++ {
				for rankIdx := 1; rankIdx <= 4; rankIdx++ {
					strRankIdx := strconv.Itoa(rankIdx)
					retPaipus, err := utils.QueryPaipuByStatus(retVips[i].Name, strRankIdx, utils.STATUS_NEW, timeIndex, 10)
					if err != nil {
						log.Println("QueryPaipuByStatus fail", err)
						continue
					}

					if len(retPaipus) > 0 {
						ptVary := 0
						zyVary := 0
						totalVary := 0
						dealVary := 0
						var paipuIds []int

						for j := 0; j < len(retPaipus); j++ {
							paipuIds = append(paipuIds, retPaipus[j].Id)
							ratePt, rateZy, err := utils.GetRateZhuyiByMode(retPaipus[j].Rate)
							if err != nil {
								log.Println("GetRateZhuyiByMode fail", err)
								continue
							}

							arrZy := []int{retPaipus[j].Zy_1, retPaipus[j].Zy_2, retPaipus[j].Zy_3, retPaipus[j].Zy_4}
							arrPt := []int{retPaipus[j].Pt_1, retPaipus[j].Pt_2, retPaipus[j].Pt_3, retPaipus[j].Pt_4}

							ptVary += arrPt[rankIdx-1]
							zyVary += arrZy[rankIdx-1]
							totalVary += int((float32(arrPt[rankIdx-1])/1000 + float32(arrZy[rankIdx-1]*rateZy)) * float32(ratePt))
							dealVary += 1
						}

						// update career todo dev
						log.Println("UpdatePlayerCareer", retVips[i].Name, dealVary)
						err = utils.UpdatePlayerCareer(retVips[i].Name, timeNow, paipuIds, ptVary, zyVary, totalVary, dealVary, strRankIdx)
						if err != nil {
							log.Println("UpdatePlayerCareer fail", err)
							continue
						}
					}
				}
			}
		}
	}
}
