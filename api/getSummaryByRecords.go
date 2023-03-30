package api

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/message"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"strconv"
	"strings"
	"time"
)

type Player struct {
	Nickname   string
	Seat       uint32
	TotalPoint int32
	Zhuyi      int
	Sum        string
}

func GetSummaryByRecords(records []string, mode string)  (map[string]*Player, []map[string]string, []map[string]string ,error){

	// validate
	uuids, err := utils.GetUuidByRecordUrl(records)
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, err
	}

	ratePt, rateZhuyi, err := utils.GetRateZhuyiByMode(mode)
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, err
	}

	// login
	mSoul, err := client.New()
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, errors.New("client fail")
	}

	n, p := utils.GetMajsoulBot()
	rspLogin, err := mSoul.Login(n,p)
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, errors.New("client 1 fail")
	}

	if rspLogin.Error != nil {
		fmt.Println(rspLogin.Error)
		return nil, nil, nil, errors.New("client 2 fail")
	}

	// get details
	reqInfo := message.ReqGameRecordsDetail{
		UuidList: uuids,
	}
	rspRecords, err := mSoul.FetchGameRecordsDetail(mSoul.Ctx, &reqInfo)
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, errors.New("detail fail")
	}

	// get record
	mapUuidByte := map[string][]byte{}

	for _,oneUuid := range uuids {
		reqPaipu := message.ReqGameRecord{
			GameUuid: oneUuid,
			ClientVersionString: mSoul.Version.Web(),
		}
		resPaipu, err := mSoul.FetchGameRecord(mSoul.Ctx, &reqPaipu)
		if err != nil {
			fmt.Println(err)
			return nil, nil, nil, errors.New("record fail")
		}

		data := resPaipu.Data
		if len(data) == 0 {
			fmt.Println(err)
			return nil, nil, nil, errors.New("record data fail")
		}

		mapUuidByte[oneUuid] = data
		time.Sleep(100)
	}

	// logout
	reqLogout := message.ReqLogout{}
	rspLogout, err := mSoul.Logout(mSoul.Ctx, &reqLogout)
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, errors.New("out 1 fail")
	}

	if rspLogout.Error != nil {
		fmt.Println(rspLogout.Error)
		return nil, nil, nil, errors.New("out 2 fail")
	}

	// analyze
	mapPlayerInfo := map[string]*Player{}
	mapPlayerIdx := make(map[string]string)
	mapIdxPlayer := make(map[string]string)
	mapUuidInfo := map[string][]*Player{}
	var ptRows []map[string]string
	var zhuyiRows []map[string]string
	idxPtRow := 0
	idxZhuyiRow := 0

	// init idx name
	for _,oneAccount := range rspRecords.RecordList[0].Accounts {
		mapPlayerIdx[oneAccount.Nickname] = "p" + strconv.Itoa(int(oneAccount.Seat))
		mapIdxPlayer[strconv.Itoa(int(oneAccount.Seat))] = oneAccount.Nickname
		mapPlayerInfo[oneAccount.Nickname] = &Player{
			Nickname: oneAccount.Nickname,
			Seat:     oneAccount.Seat,
		}
	}

	// init uuid seat name
	for _,oneRecord := range rspRecords.RecordList {
		for _, oneAccount := range oneRecord.Accounts {
			mapUuidInfo[oneRecord.Uuid] = append(mapUuidInfo[oneRecord.Uuid], &Player{
				Nickname: oneAccount.Nickname,
				Seat:     oneAccount.Seat,
			})
		}
	}

	// record pt
	for _,oneRecord := range rspRecords.RecordList {
		var onePtRow = make(map[string]string)
		idxPtRow += 1

		for _,oneResultPlayer := range oneRecord.Result.Players {
			for _,oneInfo := range mapUuidInfo[oneRecord.Uuid] {
				if oneInfo.Seat == oneResultPlayer.Seat {
					onePtRow[mapPlayerIdx[oneInfo.Nickname]] = fmt.Sprintf("%.1f", float32(oneResultPlayer.TotalPoint) / 1000)

					mapPlayerInfo[oneInfo.Nickname].TotalPoint += oneResultPlayer.TotalPoint
					break
				}
			}
		}
		onePtRow["desc"] = fmt.Sprintf("第%d局",idxPtRow)
		ptRows = append(ptRows, onePtRow)
	}

	// record zhuyi
	for oneUuid,oneByteContent := range mapUuidByte {
		idxZhuyiRow += 1
		var oneZhuyiRow = make(map[string]string)
		oneZhuyiRow["desc"] = fmt.Sprintf("第%d局", idxZhuyiRow)
		zhuyiRows = append(zhuyiRows, oneZhuyiRow)

		tmpRecord := &message.GameDetailRecords{
			Records: nil,
			Version: 0,
			Actions: nil,
			Bar:     nil,
		}

		// skip head bytes between 25 and 27
		for i:=25;i<=27;i++{
			tmpData := oneByteContent[i:]
			err = proto.Unmarshal(tmpData, tmpRecord)
			if err != nil {
				continue
			} else {
				break
			}
		}
		if err != nil {
			fmt.Println(err)
			return nil, nil, nil, errors.New("record byte fail")
		}

		for _, oneAction := range tmpRecord.Actions {
			oneAction := oneAction

			if oneAction.Result != nil {
				tmpHule := &message.RecordHule{
					Hules:       nil,
					OldScores:   nil,
					DeltaScores: nil,
					WaitTimeout: 0,
					Scores:      nil,
					Gameend:     nil,
					Doras:       nil,
					Muyu:        nil,
					Baopai:      0,
				}

				tmpData3 := oneAction.Result[19:]
				err = proto.Unmarshal(tmpData3, tmpHule)
				if err != nil || tmpHule.Hules == nil {
					continue
				} else {
					for _,v := range tmpHule.Hules {
						var tmpZ uint32 = 0
						var tmpZimo = false
						var tmpMing = false
						var oneZhuyiRow = make(map[string]string)
						oneDesc := ""

						// zimo
						if v.Zimo {
							tmpZimo = true
						}

						// ming
						for _,s := range v.Ming {
							if strings.Index(s, "shunzi") != -1 || strings.Index(s, "kezi") != -1 || strings.Index(s,"minggang") != -1 {
								tmpMing = true
								break
							}
						}

						// calc zhuyi
						if v.Yiman {
							oneDesc += "役满 "

							if v.Zimo {
								tmpZ += 5 * uint32(len(v.Fans))
							} else {
								tmpZ += 10 * uint32(len(v.Fans))
							}

							// todo yiman yifa baopai,  yiman with as
							var arrZHand []string
							arrZHand = append(arrZHand, "0s")
							arrZHand = append(arrZHand, "0m")
							arrZHand = append(arrZHand, "0p")
							for _,d := range v.LiDoras {
								oneZhuyiHand, err := utils.GetZhuyiHandByLi(d)
								if err != nil {
									fmt.Println(err)
									return nil, nil, nil, errors.New("zhuyi hand fail")
								}
								arrZHand = append(arrZHand, oneZhuyiHand)
							}

							for _,d := range v.Hand {
								for _,h := range arrZHand {
									if d == h {
										tmpZ += 1
									}

									if (h == "5p" && d == "0p") || (h == "5s" && d == "0s") || (h == "5m" && d == "0m") {
										tmpZ += 1
									}
								}
							}

							for _,d := range v.Ming {
								if strings.Index(d, "angang") != -1 {
									for _,h := range arrZHand {
										if strings.Index(d, h) != -1 {
											tmpZ += 4
										}
									}
								}
							}

						} else {
							for _,v2 := range v.Fans {
								// li and yifa
								if v2.Id == 30 {
									tmpZ += v2.Val
									oneDesc += "一发 "
								}

								if v2.Id == 33 && v2.Val > 0 {
									tmpZ += v2.Val
									oneDesc += fmt.Sprintf("里%d ", v2.Val)
								}

								// aka
								if v2.Id == 32 && tmpMing == false {
									if v2.Val == 3 {
										tmpZ += 5
										oneDesc += "AS "
									} else {
										tmpZ += v2.Val
										oneDesc += fmt.Sprintf("赤%d ", v2.Val)
									}
								}
							}
						}

						// dec zhuyi
						for i:=0;i<len(tmpHule.DeltaScores);i++ {
							if tmpHule.DeltaScores[i] < 0 {
								for _,oneInfo := range mapUuidInfo[oneUuid] {
									if i == int(oneInfo.Seat) && 0 != tmpZ{
										mapPlayerInfo[oneInfo.Nickname].Zhuyi -= int(tmpZ)
										oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = "-" + strconv.Itoa(int(tmpZ))
										break
									}
								}
							}
						}

						// inc zhuyi
						if tmpZimo {
							for _,oneInfo := range mapUuidInfo[oneUuid] {
								if v.Seat == oneInfo.Seat && 0 != tmpZ{
									mapPlayerInfo[oneInfo.Nickname].Zhuyi += int(tmpZ) * 3
									oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = strconv.Itoa(int(tmpZ)*3)
									break
								}
							}
						} else {
							for _,oneInfo := range mapUuidInfo[oneUuid] {
								if v.Seat == oneInfo.Seat && 0 != tmpZ{
									mapPlayerInfo[oneInfo.Nickname].Zhuyi += int(tmpZ)
									oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = strconv.Itoa(int(tmpZ))
									break
								}
							}
						}

						if len(oneZhuyiRow) > 0 {
							oneZhuyiRow["desc"] = oneDesc
							zhuyiRows = append(zhuyiRows, oneZhuyiRow)
						}
					}
				}
			}
		}
	}

	// calc sum
	for _,v := range mapPlayerInfo {
		money := (float32(v.TotalPoint) / 1000 + float32(v.Zhuyi) * float32(rateZhuyi)) * float32(ratePt)
		v.Sum = fmt.Sprintf("%.2f", money)
		//fmt.Println(v.seat, v.nickname, v.totalPoint, v.zhuyi, v.sum)	// test done
	}

	//fmt.Println(mapIdxPlayer)	// test done

	/*
	for _,v := range ptRows {
		fmt.Println(v)	// test done
	}

	for _,v := range zhuyiRows {
		fmt.Println(v)	// test
	}

	 */

	return mapPlayerInfo, ptRows, zhuyiRows ,nil
}
