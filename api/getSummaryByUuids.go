package api

import (
	"errors"
	"fmt"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/message"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"strconv"
	"strings"
)

// todo test
func GetSummaryByUuids(uuids []string, ratePt int, rateZhuyi int) (map[string]*Player, []map[string]string, []map[string]string, error) {

	// login
	mSoul, err := client.New()
	if err != nil {
		log.Println("new client fail", err)
		return nil, nil, nil, errors.New("client fail")
	}

	n, p := utils.GetMajsoulBot()
	rspLogin, err := mSoul.Login(n, p)
	if err != nil {
		log.Println("login fail", err)
		return nil, nil, nil, errors.New("client 1 fail")
	}

	if rspLogin.Error != nil {
		log.Println("login err", rspLogin.Error)
		return nil, nil, nil, errors.New("client 2 fail")
	}

	// logout
	defer func() {
		reqLogout := message.ReqLogout{}
		rspLogout, err := mSoul.Logout(mSoul.Ctx, &reqLogout)
		if err != nil {
			log.Println(err)
		}

		if rspLogout.Error != nil {
			log.Println(rspLogout.Error)
		}
	}()

	// get details
	reqInfo := message.ReqGameRecordsDetail{
		UuidList: uuids,
	}
	rspRecords, err := mSoul.FetchGameRecordsDetail(mSoul.Ctx, &reqInfo)
	if err != nil {
		log.Println("detail fail", err)
		return nil, nil, nil, errors.New("detail fail")
	}

	// get record
	type result struct {
		s  string
		bs []byte
	}
	var mapUuidBytes = make(map[string][]byte)
	var chanResult = make(chan result, 12)
	defer close(chanResult)

	for _, oneUuid := range uuids {
		go func(uuid string) {
			reqPaipu := message.ReqGameRecord{
				GameUuid:            uuid,
				ClientVersionString: mSoul.Version.Web(),
			}
			resPaipu, err := mSoul.FetchGameRecord(mSoul.Ctx, &reqPaipu)
			if err != nil {
				log.Println("paipu fail", err)
				chanResult <- result{
					s:  uuid,
					bs: nil,
				}
			} else {
				chanResult <- result{
					s:  uuid,
					bs: resPaipu.Data,
				}
			}
		}(oneUuid)
	}

	for i := 0; i < len(uuids); i++ {
		oneResult := <-chanResult
		if len(oneResult.bs) == 0 {
			log.Println("paipu empty" + oneResult.s)
			return nil, nil, nil, errors.New("paipu fail")
		}
		mapUuidBytes[oneResult.s] = oneResult.bs
	}

	// analyze todo

	mapPlayerInfo := map[string]*Player{}
	mapPlayerIdx := make(map[string]string)
	mapIdxPlayer := make(map[string]string)
	mapUuidInfo := map[string][]*Player{}
	var ptRows []map[string]string
	var zhuyiRows []map[string]string

	// init uuid seat nickname
	for _, oneAccount := range rspRecords.RecordList[0].Accounts {
		oneNickName := strings.TrimSpace(oneAccount.Nickname)
		oneSeat := strconv.Itoa(int(oneAccount.Seat))

		if len(oneNickName) == 0 {
			log.Println("nickname fail")
			return nil, nil, nil, errors.New("nickname fail")
		}

		mapPlayerIdx[oneNickName] = "p" + oneSeat
		mapIdxPlayer[oneSeat] = oneNickName
		mapPlayerInfo[oneNickName] = &Player{
			Nickname: oneNickName,
			Seat:     oneAccount.Seat,
		}
	}

	for _, oneRecord := range rspRecords.RecordList {
		for _, oneAccount := range oneRecord.Accounts {
			oneNickName := strings.TrimSpace(oneAccount.Nickname)

			if len(oneNickName) == 0 {
				log.Println("nickname fail")
				return nil, nil, nil, errors.New("nickname fail")
			}

			mapUuidInfo[oneRecord.Uuid] = append(mapUuidInfo[oneRecord.Uuid], &Player{
				Nickname: oneNickName,
				Seat:     oneAccount.Seat,
			})
		}
	}

	// record pt todo test uuid equal
	for _, oneUuid := range uuids {
		for _, oneRecord := range rspRecords.RecordList {
			if oneUuid == oneRecord.Uuid {
				var onePtRow = make(map[string]string)

				for _, oneResultPlayer := range oneRecord.Result.Players {
					for _, oneInfo := range mapUuidInfo[oneRecord.Uuid] {
						if oneInfo.Seat == oneResultPlayer.Seat {
							onePtRow[mapPlayerIdx[oneInfo.Nickname]] = fmt.Sprintf("%.1f", float32(oneResultPlayer.TotalPoint)/1000)

							mapPlayerInfo[oneInfo.Nickname].TotalPoint += oneResultPlayer.TotalPoint
							break
						}
					}
				}
				onePtRow["desc"] = fmt.Sprintf("第%d局", len(ptRows)+1)
				ptRows = append(ptRows, onePtRow)
				break
			}
		}
	}

	// record zhuyi
	for _, oneUuid := range uuids {
		oneDetailRecord := &message.GameDetailRecords{}
		err = utils.GetRecordFromBytes(mapUuidBytes[oneUuid], oneDetailRecord, "GameDetailRecords")
		if err != nil {
			log.Println("analyze fail", err)
			return nil, nil, nil, err
		}

		for _, oneAction := range oneDetailRecord.Actions {
			if oneAction.Result != nil {
				dictSeatYifa := [4]int{}

				// discard lizhi yifa
				oneDiscard := &message.RecordDiscardTile{}
				err = utils.GetRecordFromBytes(oneAction.Result, oneDiscard, "RecordDiscardTile")
				if err != nil {
				} else {
					if oneDiscard.IsLiqi {
						dictSeatYifa[oneDiscard.Seat] = 1
					} else {
						dictSeatYifa[oneDiscard.Seat] = 0
					}
					continue
				}

				// chi peng gang cancal yifa
				oneChiPeng := &message.RecordChiPengGang{}
				oneAnGang := &message.RecordAnGangAddGang{}
				errChiPeng := utils.GetRecordFromBytes(oneAction.Result, oneChiPeng, "RecordChiPengGang")
				errGang := utils.GetRecordFromBytes(oneAction.Result, oneAnGang, "RecordAnGangAddGang")

				if errChiPeng == nil || errGang == nil {
					for i := 0; i < len(dictSeatYifa); i++ {
						dictSeatYifa[i] = 0
					}
					continue
				}

				// hule
				oneHule := &message.RecordHule{}
				err = utils.GetRecordFromBytes(oneAction.Result, oneHule, "RecordHule")
				if err != nil || oneHule.Hules == nil {
				} else {
					for _, vHule := range oneHule.Hules {
						var cntZ uint32 = 0
						desc := ""
						bolMing := false

						var arrAnGang []string
						var oneZhuyiRow = make(map[string]string)

						// ming
						for _, vMing := range vHule.Ming {
							if strings.Index(vMing, "shunzi") != -1 || strings.Index(vMing, "kezi") != -1 || strings.Index(vMing, "minggang") != -1 {
								bolMing = true
							} else if strings.Index(vMing, "angang") != -1 {
								arrAnGang = append(arrAnGang, utils.GetHandFromAnGang(vMing)...)
							}
						}

						// calc yiman zhuyi
						if vHule.Yiman {
							desc += "役满 "

							// yiman basic zhuyi todo baopai
							if vHule.Zimo {
								cntZ += 5 * uint32(len(vHule.Fans))
							} else {
								cntZ += 10 * uint32(len(vHule.Fans))
							}

							// hand zhuyi todo yifa
							var arrHand []string
							arrHand = append(arrHand, vHule.Hand...)
							arrHand = append(arrHand, arrAnGang...)
							arrHand = append(arrHand, vHule.HuTile)

							oneCntZ, oneDesc, err := utils.GetZhuyiByHandAndLi(arrHand, vHule.LiDoras, dictSeatYifa[vHule.Seat] == 1)
							if err != nil {
								return nil, nil, nil, err
							}

							cntZ += uint32(oneCntZ)
							desc += oneDesc

							log.Println("yiman hand ", cntZ, desc, arrHand, vHule.LiDoras, dictSeatYifa[vHule.Seat] == 1)

							// calc normal zhuyi
						} else {
							for _, v2 := range vHule.Fans {
								// yifa
								if v2.Id == 30 {
									cntZ += v2.Val
									desc += "一发 "
								}

								// li
								if v2.Id == 33 && v2.Val > 0 {
									cntZ += v2.Val
									desc += fmt.Sprintf("里%d ", v2.Val)
								}

								// aka
								if v2.Id == 32 && bolMing == false {
									if v2.Val == 3 {
										cntZ += 5
										desc += "AS "
									} else {
										cntZ += v2.Val
										desc += fmt.Sprintf("赤%d ", v2.Val)
									}
								}
							}
						}

						log.Println("desc and zhuyi ", desc, cntZ)

						// dec zhuyi
						for i := 0; i < len(oneHule.DeltaScores); i++ {
							if oneHule.DeltaScores[i] < 0 {
								for _, oneInfo := range mapUuidInfo[oneUuid] {
									if i == int(oneInfo.Seat) && 0 != cntZ {
										mapPlayerInfo[oneInfo.Nickname].Zhuyi -= int(cntZ)
										oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = "-" + strconv.Itoa(int(cntZ))
										break
									}
								}
							}
						}

						// inc zhuyi
						if vHule.Zimo {
							for _, oneInfo := range mapUuidInfo[oneUuid] {
								if vHule.Seat == oneInfo.Seat && 0 != cntZ {
									mapPlayerInfo[oneInfo.Nickname].Zhuyi += int(cntZ) * 3
									oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = strconv.Itoa(int(cntZ) * 3)
									break
								}
							}
						} else {
							for _, oneInfo := range mapUuidInfo[oneUuid] {
								if vHule.Seat == oneInfo.Seat && 0 != cntZ {
									mapPlayerInfo[oneInfo.Nickname].Zhuyi += int(cntZ)
									oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = strconv.Itoa(int(cntZ))
									break
								}
							}
						}

						if len(oneZhuyiRow) > 0 {
							oneZhuyiRow["desc"] = fmt.Sprintf("第%d局 %s", len(zhuyiRows)+1, desc)
							zhuyiRows = append(zhuyiRows, oneZhuyiRow)
						}
					}
				}
			}
		}
	}

	// calc total
	for _, v := range mapPlayerInfo {
		money := (float32(v.TotalPoint)/1000 + float32(v.Zhuyi)*float32(rateZhuyi)) * float32(ratePt)
		v.Sum = fmt.Sprintf("%.2f", money)
	}

	return mapPlayerInfo, ptRows, zhuyiRows, nil
}
