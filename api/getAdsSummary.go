package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/message"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutexAds sync.Mutex

type StMajsoulLogin struct {
	err    error
	acc    string
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
	Pls        [4]string
	Totals     [4]int
	Pts        [4]int
	Zys        [4]int
	ZyDetails  [][]string
	ContestUid string
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

/**

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
	if !mu.TryLock() {
		return nil, errors.New("请求排队中，请稍后提交")
	}
	defer mutexAds.Unlock()
	time.Sleep(time.Second) // todo dev write in main

	var chanLogin = make(chan StMajsoulLogin, 1)
	var chanRecord = make(chan StMajsoulRecord, 1)
	defer close(chanLogin)
	defer close(chanRecord)

	bolChanFail := false
	var retBots []StMajsoulLogin
	var retMajsoulRecords []StMajsoulRecord
	sizeBots := min(len(utils.ConfigMajsoulBot.Acc), len(uuids))

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

	// chan meta and detail
	for i := 0; i < len(uuids); i++ {
		oneUuid := uuids[i]
		idxBot := i % sizeBots

		if bolChanFail {
			break
		} else if i > 0 && idxBot == 0 {
			for j := 0; j < sizeBots; j++ {
				oneRetRecord := <-chanRecord
				retMajsoulRecords = append(retMajsoulRecords, oneRetRecord)

				if oneRetRecord.err != nil {
					bolChanFail = true
					log.Println("getMajsoulRecordMetaDetail fail", oneRetRecord.uuid, oneRetRecord.err)
				}
			}
		} else {
			go getMajsoulRecordMetaDetail(chanRecord, oneUuid, retBots[idxBot])
		}
	}

	if bolChanFail {
		return nil, errors.New("内部错误")
	}

	// build map calc pt and zy
	mapUuid2Calc := map[string]*StCalc{}
	_, rateZy, err := utils.GetRateZhuyiByMode(mode)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for i := 0; i < len(uuids); i++ {
		oneCalc := StCalc{Uuid: uuids[i], Mode: mode}

		for j := 0; j < len(retMajsoulRecords); j++ {
			if uuids[i] == retMajsoulRecords[j].uuid {
				// seat nickname
				for _, oneAccount := range retMajsoulRecords[j].detail.RecordList[0].Accounts {
					oneNickName := strings.TrimSpace(oneAccount.Nickname)
					oneSeat := int(oneAccount.Seat)

					if len(oneNickName) == 0 {
						log.Println("nickname fail", uuids[i])
						return nil, errors.New("nickname fail")
					}

					oneCalc.Pls[oneSeat] = oneNickName
				}

				// contestUid
				oneCalc.ContestUid = strconv.Itoa(int(retMajsoulRecords[j].detail.RecordList[0].Config.Meta.ContestUid))

				// pt
				for _, oneRetPlayer := range retMajsoulRecords[j].detail.RecordList[0].Result.Players {
					oneSeat := int(oneRetPlayer.Seat)
					oneCalc.Pts[oneSeat] += int(oneRetPlayer.TotalPoint)
				}

				// zy todo dev test
				if rateZy > 0 {

				}

				break
			}
		}

		mapUuid2Calc[uuids[i]] = &oneCalc
	}

	// save record

	return nil, nil
}

// todo dev test
func loginMajsoul(chanLogin chan StMajsoulLogin, oneAcc string, onePwd string) {
	ret := StMajsoulLogin{
		acc: oneAcc,
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

func checkIfRecord(mode string, contestUid string) bool {
	// check record config mode
	for i := 0; i < len(utils.ConfigMode.RecordModeList); i++ {
		if mode == utils.ConfigMode.RecordModeList[i] {
			return true
		}
	}

	// check 503 contestUid
	for i := 0; i < len(utils.ConfigMode.RecordContestIds); i++ {
		if contestUid == utils.ConfigMode.RecordContestIds[i] {
			return true
		}
	}

	return false
}
