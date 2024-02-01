package api

import (
	"errors"
	"fmt"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/message"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

var muPt sync.Mutex

var objMajsoul *client.Majsoul
var hasInit = false
var readCount = 0

func getMajsoulRecordByUuid(chanRes chan r, uuid string, account string, p string) (*message.ResGameRecordsDetail, []byte, error) {

	// login
	if !hasInit {
		hasInit = true
		readCount += 1
		tmpMajSoul, err := client.New()
		if err != nil {
			log.Println("new client fail", err)
			chanRes <- r{
				uuid:   uuid,
				bs:     nil,
				detail: nil,
			}
			return nil, nil, errors.New("client fail")
		}

		rspLogin, err := tmpMajSoul.Login(account, p)
		if err != nil {
			log.Println("login fail", err)
			chanRes <- r{
				uuid:   uuid,
				bs:     nil,
				detail: nil,
			}
			return nil, nil, errors.New("client login fail")
		}

		if rspLogin.Error != nil {
			log.Println("login err", rspLogin.Error)
			chanRes <- r{
				uuid:   uuid,
				bs:     nil,
				detail: nil,
			}
			return nil, nil, errors.New("client login err")
		}

		objMajsoul = tmpMajSoul
	}

	// get details
	reqInfo := message.ReqGameRecordsDetail{
		UuidList: []string{uuid},
	}
	rspRecords, err := objMajsoul.FetchGameRecordsDetail(objMajsoul.Ctx, &reqInfo)
	if err != nil {
		log.Println("detail fail", err)
		chanRes <- r{
			uuid:   uuid,
			bs:     nil,
			detail: nil,
		}
		return nil, nil, errors.New("detail fail")
	}
	log.Println("FetchGameRecordsDetail ok", uuid)

	// get game record
	reqPaipu := message.ReqGameRecord{
		GameUuid:            uuid,
		ClientVersionString: objMajsoul.Version.Web(),
	}
	resPaipu, err := objMajsoul.FetchGameRecord(objMajsoul.Ctx, &reqPaipu)
	if err != nil {
		log.Println("paipu fail", err)
		chanRes <- r{
			uuid:   uuid,
			bs:     nil,
			detail: nil,
		}
		return nil, nil, errors.New("record fail")
	}

	if len(resPaipu.Data) == 0 {
		if len(resPaipu.DataUrl) != 0 {
			log.Println("data url", resPaipu.DataUrl)
		}

		log.Println("paipu data fail", err)
		chanRes <- r{
			uuid:   uuid,
			bs:     nil,
			detail: nil,
		}
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

// add one uuid
func AddRecordByUuid(uuid string) error {

	muPt.Lock()
	defer muPt.Unlock()
	var chanRes = make(chan r, 1)
	defer close(chanRes)

	// get record by uuid
	acc, pwd := utils.GetMajSoulBotByIdx(len(utils.ConfigMajsoulBot.Acc) - 1)

	_, _, err := getMajsoulRecordByUuid(chanRes, uuid, acc, pwd)
	if err != nil {
		log.Println("GetMajsoulRecordByUuid fail ", uuid)
		return err
	}

	tmpRes := <-chanRes
	if tmpRes.bs == nil {
		log.Println("majsoul record fail", tmpRes.uuid)
		return errors.New("majsoul record fail " + uuid)
	}

	// build sql data
	oneRecord := tmpRes.detail.RecordList[0]
	timestampStart := time.Unix(int64(oneRecord.StartTime), 0)
	playerCount := len(oneRecord.Accounts)

	var players []Player
	for i := 0; i < 4; i++ {
		onePlayer := Player{}
		players = append(players, onePlayer)
	}

	for i := 0; i < len(oneRecord.Accounts); i++ {
		players[i].Nickname = oneRecord.Accounts[i].Nickname
		players[i].Seat = oneRecord.Accounts[i].Seat
	}

	for i := 0; i < len(oneRecord.Result.Players); i++ {
		for j := 0; j < len(players); j++ {
			if oneRecord.Result.Players[i].Seat == players[j].Seat {
				players[j].TotalPoint = oneRecord.Result.Players[i].TotalPoint
				break
			}
		}
	}

	sort.Sort(Players(players))

	// zhuyi
	rateZhuyi := 3
	uuids := []string{uuid}

	if rateZhuyi > 0 {
		for _, oneUuid := range uuids {

			oneDetailRecord := &message.GameDetailRecords{}
			err = utils.GetRecordFromBytes(tmpRes.bs, oneDetailRecord, "GameDetailRecords")
			if err != nil {
				log.Println("analyze fail GameDetailRecords", err)
				return err
			}

			dictSeatYifa := [4]int{}
			dictSeatBaopai := [2]int{}
			idxSeatBaoapi := -1 // 大三元0 四喜1 其他-1
			for _, oneAction := range oneDetailRecord.Actions {
				if oneAction.Result != nil {

					// discard lizhi set yifa
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

					// chi peng gang unset yifa
					oneChiPeng := &message.RecordChiPengGang{}
					err = utils.GetRecordFromBytes(oneAction.Result, oneChiPeng, "RecordChiPengGang")
					if err != nil {
					} else {
						oneTile := oneChiPeng.Tiles[0]
						if oneTile == "5z" || oneTile == "6z" || oneTile == "7z" {
							dictSeatBaopai[0] = int(oneChiPeng.Froms[len(oneChiPeng.Froms)-1])
						} else if oneTile == "1z" || oneTile == "2z" || oneTile == "3z" || oneTile == "4z" {
							dictSeatBaopai[1] = int(oneChiPeng.Froms[len(oneChiPeng.Froms)-1])
						}

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

					// hule
					oneHule := &message.RecordHule{}
					err = utils.GetRecordFromBytes(oneAction.Result, oneHule, "RecordHule")
					if err != nil || oneHule.Hules == nil {
					} else {
						for _, vHule := range oneHule.Hules {
							//log.Println(oneHule.Hules) //test
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

							if vHule.Yiman {
								// calc yiman zhuyi basic and baopai seat
								if vHule.Baopai > 0 {
									cntZ += 15

									// 大三元id=37
									for _, oneFan := range vHule.Fans {
										if oneFan.Id == 37 {
											idxSeatBaoapi = 0
											break
										}
									}

									if idxSeatBaoapi == -1 {
										idxSeatBaoapi = 1
									}

								} else {
									if vHule.Zimo {
										cntZ += 5
									} else {
										cntZ += 10
									}
								}

								// hand zhuyi
								var arrHand []string
								arrHand = append(arrHand, vHule.Hand...)
								arrHand = append(arrHand, arrAnGang...)
								arrHand = append(arrHand, vHule.HuTile)

								oneCntZ, oneDesc, err := utils.GetZhuyiByHandAndLi(arrHand, vHule.LiDoras, dictSeatYifa[vHule.Seat] == 1, bolMing)
								if err != nil {
									log.Println("yiman hand fail", err)
									return err
								}

								cntZ += uint32(oneCntZ)
								desc += "役满"
								desc += oneDesc

								log.Println("yakuman ", oneUuid, cntZ, desc, arrHand, vHule.LiDoras, dictSeatYifa[vHule.Seat] == 1)
							} else {
								// calc normal zhuyi
								for _, v2 := range vHule.Fans {
									// yifa
									if v2.Id == 30 {
										cntZ += v2.Val
										desc += "一发"
									}

									// li
									if v2.Id == 33 && v2.Val > 0 {
										cntZ += v2.Val
										desc += fmt.Sprintf("里%d", v2.Val)
									}

									// aka
									if v2.Id == 32 && bolMing == false {
										if v2.Val == 3 {
											cntZ += 5
											desc += "AS"
										} else {
											cntZ += v2.Val
											desc += fmt.Sprintf("赤%d", v2.Val)
										}
									}
								}
							}

							if vHule.Baopai > 0 && 0 != cntZ {
								for i := 0; i < len(players); i++ {
									if idxSeatBaoapi != -1 && int(players[i].Seat) == dictSeatBaopai[idxSeatBaoapi] {
										players[i].Zhuyi -= int(cntZ)
									}
									if vHule.Seat == players[i].Seat {
										players[i].Zhuyi += int(cntZ)
									}
								}

							} else if 0 != cntZ {
								// dec zhuyi by scores
								for i := 0; i < len(oneHule.DeltaScores); i++ {
									if oneHule.DeltaScores[i] < 0 {
										for j := 0; j < len(players); j++ {
											if i == int(players[j].Seat) {
												players[j].Zhuyi -= int(cntZ)
											}
										}
									}
								}

								// inc zhuyi by seat
								if vHule.Zimo {
									for i := 0; i < len(players); i++ {
										if vHule.Seat == players[i].Seat {
											players[i].Zhuyi += int(cntZ) * (playerCount - 1)
											break
										}
									}
								} else {
									for i := 0; i < len(players); i++ {
										if vHule.Seat == players[i].Seat {
											players[i].Zhuyi += int(cntZ)
											break
										}
									}
								}
							}

							// reset
							dictSeatYifa = [4]int{}
							dictSeatBaopai = [2]int{}
							idxSeatBaoapi = -1
						}
					}
				}
			}
		}
	}

	// insert db
	var onePt = utils.Table_pt{
		Paipu_url:    uuid,
		Player_count: playerCount,
		Time_start:   timestampStart,
		Pl_1:         players[0].Nickname,
		Pl_2:         players[1].Nickname,
		Pl_3:         players[2].Nickname,
		Pl_4:         players[3].Nickname,
		Pt_1:         int(players[0].TotalPoint),
		Pt_2:         int(players[1].TotalPoint),
		Pt_3:         int(players[2].TotalPoint),
		Pt_4:         int(players[3].TotalPoint),
		Zy_1:         players[0].Zhuyi,
		Zy_2:         players[1].Zhuyi,
		Zy_3:         players[2].Zhuyi,
		Zy_4:         players[3].Zhuyi,
	}

	_, err = utils.InsertPt(onePt)
	if err != nil {
		log.Println("insert fail", err)
		return err
	}

	return nil
}

/*
	for i := 0; i < len(players); i++ {
		log.Println(players[i])
	}
	log.Println(timestampStart, playerCount)

*/
