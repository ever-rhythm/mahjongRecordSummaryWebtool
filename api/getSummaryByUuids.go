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
	"sync"
	"time"
)

const MAX_GONUM = 4

type r struct {
	uuid   string
	bs     []byte
	detail *message.ResGameRecordsDetail
}

// test done
func getMsoulRecordByUuid(chanRes chan r, uuid string, account string, p string) (*message.ResGameRecordsDetail, []byte, error) {

	// login
	mSoul, err := client.New()
	if err != nil {
		log.Println("new client fail", err)
		return nil, nil, errors.New("client fail")
	}

	rspLogin, err := mSoul.Login(account, p)
	if err != nil {
		log.Println("login fail", err)
		return nil, nil, errors.New("client 1 fail")
	}

	if rspLogin.Error != nil {
		log.Println("login err", rspLogin.Error)
		return nil, nil, errors.New("client 2 fail")
	}

	// get details
	reqInfo := message.ReqGameRecordsDetail{
		UuidList: []string{uuid},
	}
	rspRecords, err := mSoul.FetchGameRecordsDetail(mSoul.Ctx, &reqInfo)
	if err != nil {
		log.Println("detail fail", err)
		return nil, nil, errors.New("detail fail")
	}
	log.Println("FetchGameRecordsDetail ok ", uuid)

	// get game record
	reqPaipu := message.ReqGameRecord{
		GameUuid:            uuid,
		ClientVersionString: mSoul.Version.Web(),
	}
	resPaipu, err := mSoul.FetchGameRecord(mSoul.Ctx, &reqPaipu)
	if err != nil {
		log.Println("paipu fail", err)
		return nil, nil, errors.New("record fail")
	}

	if len(resPaipu.Data) == 0 {
		log.Println("paipu data fail", err)
		return nil, nil, errors.New("record data fail")
	}

	log.Println("FetchGameRecord ok ", uuid)

	chanRes <- r{
		uuid:   uuid,
		bs:     resPaipu.Data,
		detail: rspRecords,
	}

	return nil, nil, nil
}

// test done
func GetSummaryByUuids(uuids []string, ratePt int, rateZhuyi int) (map[string]*Player, []map[string]string, []map[string]string, error) {

	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	var err error
	var chanRes = make(chan r, 1)
	defer close(chanRes)

	var mapUuidBytes = make(map[string][]byte)
	rspRecords := message.ResGameRecordsDetail{}

	for i := 0; i < len(uuids); i++ {
		idx := i % MAX_GONUM
		tmpUuid := uuids[i]
		account, p := utils.GetMajSoulBotByIdx(idx)
		go getMsoulRecordByUuid(chanRes, tmpUuid, account, p)

		if i > 0 && idx == 0 {
			time.Sleep(4 * time.Second)
		}
	}

	for i := 0; i < len(uuids); i++ {
		tmpRes := <-chanRes
		if len(tmpRes.uuid) == 0 {
			log.Println("msoul record fail")
			return nil, nil, nil, errors.New("msoul record fail")
		}

		mapUuidBytes[tmpRes.uuid] = tmpRes.bs
		rspRecords.RecordList = append(rspRecords.RecordList, tmpRes.detail.RecordList...)
	}

	//return nil, nil, nil, errors.New("test") //test
	// analyze
	mapPlayerInfo := map[string]*Player{}
	mapPlayerIdx := make(map[string]string)
	mapIdxPlayer := make(map[string]string)
	mapUuidInfo := map[string][]*Player{}
	var ptRows []map[string]string
	var zhuyiRows []map[string]string
	idxZhuyi := 0

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

			_, b := mapPlayerIdx[oneNickName]
			if !b {
				return nil, nil, nil, errors.New("different 4 player name " + oneNickName)
			}

			mapUuidInfo[oneRecord.Uuid] = append(mapUuidInfo[oneRecord.Uuid], &Player{
				Nickname: oneNickName,
				Seat:     oneAccount.Seat,
			})
		}
	}

	// record pt test uuid equal done
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
		idxZhuyi += 1
		onePreDesc := fmt.Sprintf("第%d局 ", idxZhuyi)

		oneDetailRecord := &message.GameDetailRecords{}
		err = utils.GetRecordFromBytes(mapUuidBytes[oneUuid], oneDetailRecord, "GameDetailRecords")
		if err != nil {
			log.Println("analyze fail GameDetailRecords", err)
			return nil, nil, nil, err
		}

		dictSeatYifa := [4]int{}
		for _, oneAction := range oneDetailRecord.Actions {
			if oneAction.Result != nil {

				// discard lizhi yifa
				oneDiscard := &message.RecordDiscardTile{}
				err = utils.GetRecordFromBytes(oneAction.Result, oneDiscard, "RecordDiscardTile")
				if err != nil {
				} else {
					if oneDiscard.IsLiqi {
						//fmt.Println("liqi ", oneDiscard.Seat) //test
						dictSeatYifa[oneDiscard.Seat] = 1
					} else {
						dictSeatYifa[oneDiscard.Seat] = 0
					}
					continue
				}

				// chi peng gang cancal yifa
				oneChiPeng := &message.RecordChiPengGang{}
				err = utils.GetRecordFromBytes(oneAction.Result, oneChiPeng, "RecordChiPengGang")
				if err != nil {
				} else {
					dictSeatYifa = [4]int{}
					continue
				}

				oneAnGang := &message.RecordAnGangAddGang{}
				err = utils.GetRecordFromBytes(oneAction.Result, oneAnGang, "RecordAnGangAddGang")
				if err != nil {
				} else {
					dictSeatYifa = [4]int{}
					continue
				}

				//fmt.Println("dict yifa", dictSeatYifa) //test

				// hule
				oneHule := &message.RecordHule{}
				err = utils.GetRecordFromBytes(oneAction.Result, oneHule, "RecordHule")
				if err != nil || oneHule.Hules == nil {
				} else {
					for _, vHule := range oneHule.Hules {
						var cntZ uint32 = 0
						desc := ""
						bolMing := false

						var oneZhuyiRow = make(map[string]string)
						oneZhuyiRow["p0"] = " "
						oneZhuyiRow["p1"] = " "
						oneZhuyiRow["p2"] = " "
						oneZhuyiRow["p3"] = " "

						// ming
						var arrAnGang []string
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

							// yiman basic zhuyi
							// undo baopai
							if vHule.Baopai > 0 {
								cntZ += 15
							} else {
								if vHule.Zimo {
									cntZ += 5
								} else {
									cntZ += 10
								}
							}

							//fmt.Println(vHule.Baopai) //test

							// hand zhuyi test yifa done
							var arrHand []string
							arrHand = append(arrHand, vHule.Hand...)
							arrHand = append(arrHand, arrAnGang...)
							arrHand = append(arrHand, vHule.HuTile)

							//fmt.Println(dictSeatYifa, vHule.Seat) //test
							oneCntZ, oneDesc, err := utils.GetZhuyiByHandAndLi(arrHand, vHule.LiDoras, dictSeatYifa[vHule.Seat] == 1, bolMing)
							if err != nil {
								log.Println("yiman hand fail", err)
								return nil, nil, nil, err
							}

							cntZ += uint32(oneCntZ)
							desc += oneDesc
							dictSeatYifa = [4]int{}

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

						//log.Println("desc and zhuyi ", desc, cntZ)

						// dec zhuyi by scores
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

						// inc zhuyi by scores
						if vHule.Zimo && vHule.Baopai == 0 {
							for _, oneInfo := range mapUuidInfo[oneUuid] {
								if vHule.Seat == oneInfo.Seat && 0 != cntZ {
									mapPlayerInfo[oneInfo.Nickname].Zhuyi += int(cntZ) * 3
									oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = "+" + strconv.Itoa(int(cntZ)*3)
									break
								}
							}
						} else {
							for _, oneInfo := range mapUuidInfo[oneUuid] {
								if vHule.Seat == oneInfo.Seat && 0 != cntZ {
									mapPlayerInfo[oneInfo.Nickname].Zhuyi += int(cntZ)
									oneZhuyiRow[mapPlayerIdx[oneInfo.Nickname]] = "+" + strconv.Itoa(int(cntZ))
									break
								}
							}
						}

						if len(desc) > 0 {
							oneZhuyiRow["desc"] = onePreDesc + desc
							onePreDesc = ""
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

	log.Println(uuids, ptRows, zhuyiRows)
	for _, v := range mapPlayerInfo {
		log.Println(v)
	}

	return mapPlayerInfo, ptRows, zhuyiRows, nil
}

// logout
/*
	defer func() {
		//mSoul.LobbyConn.Close()

		reqLogout := message.ReqLogout{}
		rspLogout, err := mSoul.Logout(mSoul.Ctx, &reqLogout)
		if err != nil {
			log.Println(err)
		}

		if rspLogout.Error != nil {
			log.Println(rspLogout.Error)
		}
	}()

*/

// get record socket read error ?
/*
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

*/

// logout 好像没啥用 defer conn loop panic
/*
	reqLogout := message.ReqLogout{}
	rspLogout, err := mSoul.Logout(mSoul.Ctx, &reqLogout)
	if err != nil {
		log.Println(err)
	}
	if rspLogout.Error != nil {
		log.Println(rspLogout.Error)
	}
*/
