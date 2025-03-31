package api

import (
	"errors"
	"fmt"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/message"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutexAds sync.Mutex

type StMajsoulLogin struct {
	err    error
	acc    string
	pwd    string
	client *client.Majsoul
}

type StMajsoulRecord struct {
	err    error
	uuid   string
	bs     []byte
	detail *message.ResGameRecordsDetail
}

type StCalc struct {
	Uuid       string
	Mode       string
	ContestUid int
	PlayerCnt  int
	TimeStart  int

	Pls         [4]string
	Totals      [4]float32
	Pts         [4]int // format 25000
	Zys         [4]int
	ZyDescs     []string
	ZyDetailCnt [][4]int
}

type StZyHule struct {
	ZyCnt  [4]int
	ZyDesc string

	ZyLi    [4]int
	ZyAka   [4]int
	ZyYifa  [4]int
	ZyYiman [4]int
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

/** GetAdsSummary

mutex
login
get majsoul meta
get majsoul detail

build map
calc pt
calc zy

check if record
save paipu
save player
save action

*/

func GetAdsSummary(uuids []string, mode string) ([]StCalc, error) {

	// lock
	if !mutexAds.TryLock() {
		return nil, errors.New("请求排队中，请稍后提交")
	}
	defer mutexAds.Unlock()
	time.Sleep(time.Second)

	// chan start
	var chanLogin = make(chan StMajsoulLogin, 1)
	defer close(chanLogin)
	var chanRecord = make(chan StMajsoulRecord, 1)
	defer close(chanRecord)
	var retBots []StMajsoulLogin
	var retMajsoulRecords []StMajsoulRecord

	sizeBots := min(len(utils.ConfigMajsoulBot.Acc), len(uuids))
	ratePt, rateZy, err := utils.GetRateZhuyiByMode(mode)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	bolChanFail := false

	// chan login
	for i := 0; i < sizeBots; i++ {
		oneAcc, onePwd := utils.GetMajSoulBotByIdx(i)
		go loginMajsoul(chanLogin, oneAcc, onePwd)
	}

	for i := 0; i < sizeBots; i++ {
		oneRetLogin := <-chanLogin
		retBots = append(retBots, oneRetLogin)

		if oneRetLogin.err != nil {
			bolChanFail = true
			log.Println("loginMajsoul fail", oneRetLogin.acc, oneRetLogin.err)
		}
	}

	if bolChanFail {
		return nil, errors.New("内部错误")
	}

	// 试试1个client 同时请求多个，会不会有问题
	// chan meta, chan detail
	for i := 0; i < len(uuids); i++ {
		oneUuid := uuids[i]
		idxBot := i % sizeBots
		if i > 0 && idxBot == 0 {
			time.Sleep(time.Second)
		}
		go getMajsoulRecordMetaDetail(chanRecord, oneUuid, retBots[idxBot])
	}

	for i := 0; i < len(uuids); i++ {
		oneRetRecord := <-chanRecord
		retMajsoulRecords = append(retMajsoulRecords, oneRetRecord)

		if oneRetRecord.err != nil {
			bolChanFail = true
			log.Println("getMajsoulRecordMetaDetail fail", oneRetRecord.uuid, oneRetRecord.err)
		}
	}

	if bolChanFail {
		return nil, errors.New("牌谱获取失败")
	}

	// build map , calc pt and zy
	var arrStCalc []StCalc
	mapPl2Seat := make(map[string]int)

	// build mapPl2seat and check nickname
	for j := 0; j < len(retMajsoulRecords); j++ {
		for _, oneAccount := range retMajsoulRecords[j].detail.RecordList[0].Accounts {
			oneNickName := strings.TrimSpace(oneAccount.Nickname)
			oneSeat := int(oneAccount.Seat)

			if len(oneNickName) == 0 {
				log.Println("get nickname fail", retMajsoulRecords[j].uuid)
				return nil, errors.New("昵称获取失败")
			}

			if j == 0 {
				mapPl2Seat[oneNickName] = oneSeat
			} else {
				_, b := mapPl2Seat[oneNickName]
				if !b {
					log.Println("paipu nickname different", retMajsoulRecords[j].uuid, oneNickName)
					return nil, errors.New("牌谱之间昵称不同：" + oneNickName)
				}
			}
		}
	}

	// 多个牌谱 seat 不一致，save record 和 api 估计逻辑不同，api层自己汇总吧
	for i := 0; i < len(uuids); i++ {
		oneCalc := StCalc{Uuid: uuids[i], Mode: mode}

		for j := 0; j < len(retMajsoulRecords); j++ {
			if uuids[i] == retMajsoulRecords[j].uuid {
				// contestUid = ContestUid/group_id
				if int(retMajsoulRecords[j].detail.RecordList[0].Config.Meta.ContestUid) > 0 {
					oneCalc.ContestUid = int(retMajsoulRecords[j].detail.RecordList[0].Config.Meta.ContestUid)
				} else {
					tmpContestUid, _ := strconv.Atoi(mode)
					oneCalc.ContestUid = tmpContestUid
				}

				oneCalc.PlayerCnt = len(retMajsoulRecords[j].detail.RecordList[0].Accounts)
				oneCalc.TimeStart = int(retMajsoulRecords[j].detail.RecordList[0].StartTime)

				// seat nickname
				for _, oneAccount := range retMajsoulRecords[j].detail.RecordList[0].Accounts {
					oneNickName := strings.TrimSpace(oneAccount.Nickname)
					oneSeat := int(oneAccount.Seat)
					oneCalc.Pls[oneSeat] = oneNickName
				}

				// pt
				for _, oneRetPlayer := range retMajsoulRecords[j].detail.RecordList[0].Result.Players {
					oneSeat := int(oneRetPlayer.Seat)
					oneCalc.Pts[oneSeat] += int(oneRetPlayer.TotalPoint)
				}

				// zy
				if rateZy > 0 {
					retZyHule, err := calcZyByDetail(retMajsoulRecords[j].bs, uuids[i], len(retMajsoulRecords[j].detail.RecordList[0].Accounts))
					if err != nil {
						log.Println("calcZyByDetail fail", err)
						return nil, errors.New("牌谱获取失败")
					}

					for k := 0; k < len(retZyHule); k++ {
						// zy total
						for l := 0; l < 4; l++ {
							oneCalc.Zys[l] += retZyHule[k].ZyCnt[l]
						}

						// zy desc
						if len(retZyHule[k].ZyDesc) > 0 {
							oneCalc.ZyDetailCnt = append(oneCalc.ZyDetailCnt, retZyHule[k].ZyCnt)
							oneCalc.ZyDescs = append(oneCalc.ZyDescs, retZyHule[k].ZyDesc)
						}
					}
				}

				// total
				for l := 0; l < 4; l++ {
					oneCalc.Totals[l] = float32(oneCalc.Pts[l]*ratePt)/1000 + float32(oneCalc.Zys[l]*ratePt*rateZy)
				}

				break
			}
		}

		arrStCalc = append(arrStCalc, oneCalc)
	}

	// check and save record
	var tmpPls []string
	for i := 0; i < len(arrStCalc[0].Pls); i++ {
		if len(arrStCalc[0].Pls[i]) > 0 {
			tmpPls = append(tmpPls, arrStCalc[0].Pls[i])
		}
	}
	if checkIfRecord(arrStCalc[0].Mode, arrStCalc[0].ContestUid, tmpPls, utils.GetCurMonthDate()) {
		savePlAndPaipu(arrStCalc)
	}

	return arrStCalc, nil
}

// 需要排序，牌谱默认seat 是东南西北
func savePlAndPaipu(arrStCalc []StCalc) {
	if len(arrStCalc) == 0 {
		return
	}

	var batchPls []utils.TablePlayer
	var batchPaipus []utils.TablePaipu

	// build batchPls
	for i := 0; i < arrStCalc[0].PlayerCnt; i++ {
		batchPls = append(batchPls, utils.TablePlayer{
			Name:     arrStCalc[0].Pls[i],
			Group_Id: arrStCalc[0].ContestUid, // group_id = contestUid
		})
	}

	// build batchPaipus
	for _, oneCalc := range arrStCalc {
		// sort
		var players []Player
		for i := 0; i < oneCalc.PlayerCnt; i++ {
			onePlayer := Player{
				Nickname:   oneCalc.Pls[i],
				TotalPoint: int32(oneCalc.Pts[i]),
				Zhuyi:      oneCalc.Zys[i],
			}
			players = append(players, onePlayer)
		}
		sort.Sort(Players(players))

		// build
		if oneCalc.PlayerCnt == 4 {
			batchPaipus = append(batchPaipus, utils.TablePaipu{
				Paipu_Url:    oneCalc.Uuid,
				Player_Count: oneCalc.PlayerCnt,
				Time_Start:   time.Unix(int64(oneCalc.TimeStart), 0),
				Rate:         oneCalc.Mode,
				Group_Id:     arrStCalc[0].ContestUid,
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
			})
		} else if oneCalc.PlayerCnt == 3 {
			batchPaipus = append(batchPaipus, utils.TablePaipu{
				Paipu_Url:    oneCalc.Uuid,
				Player_Count: oneCalc.PlayerCnt,
				Time_Start:   time.Unix(int64(oneCalc.TimeStart), 0),
				Rate:         oneCalc.Mode,
				Group_Id:     arrStCalc[0].ContestUid,
				Pl_1:         players[0].Nickname,
				Pl_2:         players[1].Nickname,
				Pl_3:         players[2].Nickname,
				Pl_4:         "",
				Pt_1:         int(players[0].TotalPoint),
				Pt_2:         int(players[1].TotalPoint),
				Pt_3:         int(players[2].TotalPoint),
				Pt_4:         0,
				Zy_1:         players[0].Zhuyi,
				Zy_2:         players[1].Zhuyi,
				Zy_3:         players[2].Zhuyi,
				Zy_4:         0,
			})
		}

	}

	// batch insert
	log.Println("BatchInsert", batchPaipus)
	//go utils.BatchInsertPlayer(batchPls)
	go utils.BatchInsertPaipu(batchPaipus)
}

func calcZyByDetail(oneDetailBytes []byte, oneUuid string, playerCnt int) ([]StZyHule, error) {
	// build output
	var err error
	var arrZyHule []StZyHule

	// get struct from bytes
	oneDetailRecord := &message.GameDetailRecords{}
	err = utils.GetRecordFromBytes(oneDetailBytes, oneDetailRecord, "GameDetailRecords")
	if err != nil {
		log.Println("analyze fail GameDetailRecords", err)
		return nil, err
	}

	// analyze action
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
					var cntZ uint32 = 0
					desc := ""
					bolMing := false

					// build one zy hule
					oneZyHule := StZyHule{}

					// ming
					var arrAnGang []string
					for _, vMing := range vHule.Ming {
						if strings.Index(vMing, "shunzi") != -1 || strings.Index(vMing, "kezi") != -1 || strings.Index(vMing, "minggang") != -1 {
							bolMing = true
						} else if strings.Index(vMing, "angang") != -1 {
							arrAnGang = append(arrAnGang, utils.GetHandFromAnGang(vMing)...)
						}
					}

					// calc yiman zhuyi basic and baopai seat
					if vHule.Yiman {
						if vHule.Baopai > 0 {
							if vHule.Zimo {
								cntZ += 15
							} else {
								cntZ += 10
							}

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
							log.Println("yiman hand analyze fail", err)
							return nil, err
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

					// baopai
					if vHule.Baopai > 0 && 0 != cntZ {
						if idxSeatBaoapi != -1 {
							oneZyHule.ZyCnt[dictSeatBaopai[idxSeatBaoapi]] -= int(cntZ)
						}

						oneZyHule.ZyCnt[vHule.Seat] += int(cntZ)
					} else if 0 != cntZ {
						// normal
						for i := 0; i < len(oneHule.DeltaScores); i++ {
							if oneHule.DeltaScores[i] < 0 {
								oneZyHule.ZyCnt[i] -= int(cntZ)
							}
						}

						if vHule.Zimo {
							oneZyHule.ZyCnt[vHule.Seat] += int(cntZ) * (playerCnt - 1)
						} else {
							oneZyHule.ZyCnt[vHule.Seat] += int(cntZ)
						}
					}

					// output
					oneZyHule.ZyDesc = desc
					arrZyHule = append(arrZyHule, oneZyHule)

					// reset
					dictSeatYifa = [4]int{}
					dictSeatBaopai = [2]int{}
					idxSeatBaoapi = -1
				}
			}
		}
	}

	return arrZyHule, nil
}

func loginMajsoul(chanLogin chan StMajsoulLogin, oneAcc string, onePwd string) {
	ret := StMajsoulLogin{
		acc: oneAcc,
		pwd: onePwd,
	}

	objMajsoul, err := client.New()
	if err != nil {
		ret.err = err
		chanLogin <- ret
		return
	}

	rspLogin, err := objMajsoul.Login(oneAcc, onePwd)
	if err != nil || rspLogin.Error != nil {
		ret.err = errors.Join(err, errors.New(rspLogin.Error.String()))
		chanLogin <- ret
		return
	}

	ret.client = objMajsoul
	chanLogin <- ret
	return
}

func getMajsoulRecordMetaDetail(chanRecord chan StMajsoulRecord, oneUuid string, oneBot StMajsoulLogin) {
	ret := StMajsoulRecord{
		uuid: oneUuid,
	}

	reqInfo := message.ReqGameRecordsDetail{
		UuidList: []string{oneUuid},
	}

	rspRecords, err := oneBot.client.FetchGameRecordsDetail(oneBot.client.Ctx, &reqInfo)
	if err != nil {
		ret.err = err
		chanRecord <- ret
		return
	}

	reqPaipu := message.ReqGameRecord{
		GameUuid:            oneUuid,
		ClientVersionString: oneBot.client.Version.Web(),
	}

	resPaipu, err := oneBot.client.FetchGameRecord(oneBot.client.Ctx, &reqPaipu)
	if err != nil || len(resPaipu.Data) == 0 {
		ret.err = errors.Join(err, errors.New("paipu data len="+strconv.Itoa(len(resPaipu.Data))))
		chanRecord <- ret
		return
	}

	ret.detail = rspRecords
	ret.bs = resPaipu.Data
	chanRecord <- ret
	return
}

func checkIfRecord(mode string, contestUid int, pls []string, date string) bool {
	// check mode
	for i := 0; i < len(utils.ConfigMode.RecordModeList); i++ {
		if mode == utils.ConfigMode.RecordModeList[i] {
			return true
		}
	}

	// check 503/203 room contestUid
	for i := 0; i < len(utils.ConfigMode.RecordContestIds); i++ {
		if strconv.Itoa(contestUid) == utils.ConfigMode.RecordContestIds[i] {
			return true
		}
	}

	// check vip pl
	retVip, err := utils.QueryVipPlayer(contestUid, pls, date)
	if err != nil {
		log.Println("QueryVipPlayer fail", err)
		return false
	}

	if len(retVip) > 0 {
		return true
	}

	return false
}
