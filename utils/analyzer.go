package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
	"strconv"
	"strings"
	"time"
)

type configMajsoulBot struct {
	Acc []string
	Pwd []string
}

type configMode struct {
	Mode             []string
	Rate             []string
	Zy               []string
	RecordContestIds []string
	RecordMode       int
	RecordModeList   []string
	RecordSwitch     int
	MapRecordVips    map[string]int
}

var ConfigMajsoulBot = configMajsoulBot{}
var ConfigMode = configMode{}
var Salt = "mjgetzhjds"

func GetRecordFromBytes(bs []byte, st protoreflect.ProtoMessage, oneRecordName string) error {

	// find split ascii
	var idx int
	for i := 0; i < len(bs); i++ {
		if bs[i] == 16 || bs[i] == 8 || bs[i] == 1 {
			idx = i
			break
		}
	}

	// fast filter
	if strings.Index(string(bs[:idx+1]), oneRecordName) == -1 {
		return errors.New("record name not match")
	}

	// fix magic ascii
	if oneRecordName == "GameDetailRecords" {
		var err error
		for i := 27; i >= 25; i-- {
			err = proto.Unmarshal(bs[i:], st)
			if err != nil {
				continue
			} else {
				break
			}
		}

		if err != nil {
			return err
		}
	} else {
		if oneRecordName == "RecordHule" {
			idx = 19 // magic
		} else if oneRecordName == "RecordAnGangAddGang" {
			idx += 1
		}

		err := proto.Unmarshal(bs[idx:], st)
		if err != nil {
			return err
		}
	}

	return nil
}

// undo test 2 ge angang
func GetHandFromAnGang(s string) []string {
	tmpLeft := strings.Index(s, "(")
	tmpRight := strings.Index(s, ")")
	if tmpLeft != -1 && tmpRight != -1 {
		tmpHand := s[tmpLeft+1 : tmpRight]
		return strings.Split(tmpHand, ",")
	}

	return nil
}

func GetZhuyiByHandAndLi(hands []string, li []string, bolYifa bool, bolMing bool) (int, string, error) {
	cntYifa := 0
	cntLi := 0
	cntAka := 0
	desc := ""

	if !bolMing {
		// yifa
		if bolYifa {
			cntYifa += 1
			desc += "一发"
		}

		// li init
		mapLi := make(map[string]int)
		for _, oneLi := range li {
			oneLiHand, err := GetZhuyiHandByLi(oneLi)
			if err != nil {
				return 0, "", err
			}
			mapLi[oneLiHand] += 1

			if oneLiHand == "5s" {
				mapLi["0s"] += 1
			} else if oneLiHand == "5m" {
				mapLi["0m"] += 1
			} else if oneLiHand == "5p" {
				mapLi["0p"] += 1
			}
		}

		// aka init
		mapAka := make(map[string]int)
		mapAka["0s"] = 1
		mapAka["0m"] = 1
		mapAka["0p"] = 1

		for _, vHand := range hands {
			if mapAka[vHand] == 1 {
				cntAka += 1
			}

			if mapLi[vHand] == 1 {
				cntLi += mapLi[vHand]
			}
		}

		// li
		if cntLi > 0 {
			desc += fmt.Sprintf("里%d", cntLi)
		}

		// aka
		if cntAka > 0 {
			if cntAka == 3 {
				cntAka = 5
				desc += "AS"
			} else {
				desc += fmt.Sprintf("赤%d", cntAka)
			}
		}
	}

	return cntYifa + cntLi + cntAka, desc, nil
}

func GetZhuyiHandByLi(s string) (string, error) {
	if len(s) != 2 {
		return "", errors.New("invalid hand len")
	}

	num, err := strconv.Atoi(s[0:1])
	if err != nil {
		return "", err
	}
	// 0s
	if num == 0 {
		num = 5
	}

	tp := s[1:2]

	if tp == "z" {
		if num <= 4 {
			num += 1
			if num > 4 {
				num = 1
			}
		} else if num == 5 {
			num = 6
		} else if num == 6 {
			num = 7
		} else if num == 7 {
			num = 5
		}
	} else if tp == "m" || tp == "p" || tp == "s" {
		num += 1
		if num > 9 {
			num = 1
		}
	} else {
		return "", errors.New("invalid hand char")
	}

	return strconv.Itoa(num) + tp, nil
}

func GetUuidByRecordUrl(arrUrl []string) ([]string, error) {
	var arrUuid []string

	for _, oneUrl := range arrUrl {
		oneIdx := strings.Index(oneUrl, "paipu=")
		if oneIdx == -1 {
			return nil, errors.New("url err no paipu " + oneUrl)
		}

		oneLastIdx := strings.LastIndex(oneUrl, "_")
		if oneLastIdx == -1 {
			return nil, errors.New("url err no underline " + oneUrl)
		}

		cntUnderLine := strings.Count(oneUrl, "_")
		if cntUnderLine != 1 {
			return nil, errors.New("url err invalid underline " + oneUrl)
		}

		oneUuid := oneUrl[oneIdx+len("paipu=") : oneLastIdx]
		cntSep := strings.Count(oneUuid, "-")
		if cntSep != 5 {
			return nil, errors.New("url err invalid sep " + oneUrl)
		}

		arrUuid = append(arrUuid, oneUuid)
	}

	if len(arrUuid) == 0 {
		return arrUuid, errors.New("url empty list")
	}

	return arrUuid, nil
}

func GetRateZhuyiByMode(mode string) (int, int, error) {
	for i := 0; i < len(ConfigMode.Mode); i++ {
		if ConfigMode.Mode[i] == mode {
			rate, _ := strconv.Atoi(ConfigMode.Rate[i])
			zy, _ := strconv.Atoi(ConfigMode.Zy[i])
			return rate, zy, nil
		}
	}
	return 0, 0, errors.New("mode not support")
}

func GetMajSoulBotByIdx(idx int) (string, string) {
	idxMod := idx % len(ConfigMajsoulBot.Acc)
	return ConfigMajsoulBot.Acc[idxMod], ConfigMajsoulBot.Pwd[idxMod]
}

func GetCurMonthDate() string {
	return time.Now().Format("2006-01") + "-01"
}

// GetNextMonthDate 只适用于 02-01 这种格式，不然 addDate 可能跨2个月
func GetNextMonthDate(date string) (string, error) {
	if len(date) != 10 {
		return "", errors.New("err format")
	}

	tmpDate := date[:len(date)-3] + "-01"
	newDate, err := time.Parse("2006-01-02", tmpDate)
	if err != nil {
		return "", err
	}

	return newDate.AddDate(0, 1, 0).Format("2006-01-02"), nil
}

func GetMidMonthDate(date string) (string, error) {
	if len(date) != 10 {
		return "", errors.New("err format")
	}

	tmpDate := date[:len(date)-3] + "-01"
	newDate, err := time.Parse("2006-01-02", tmpDate)
	if err != nil {
		return "", err
	}

	return newDate.AddDate(0, 0, 13).Add(23 * time.Hour).Format("2006-01-02T15:04:05"), nil
}

func GetPreMonthDate(date string) (string, error) {
	if len(date) != 10 {
		return "", errors.New("err format")
	}

	tmpDate := date[:len(date)-3] + "-01"
	newDate, err := time.Parse("2006-01-02", tmpDate)
	if err != nil {
		return "", err
	}

	return newDate.AddDate(0, 0, -1).Add(23 * time.Hour).Format("2006-01-02T15:04:05"), nil
}

func GetEndMonthDate(date string) (string, error) {
	if len(date) != 10 {
		return "", errors.New("err format")
	}

	tmpDate := date[:len(date)-3] + "-01"
	newDate, err := time.Parse("2006-01-02", tmpDate)
	if err != nil {
		return "", err
	}

	return newDate.AddDate(0, 1, -1).Add(23 * time.Hour).Format("2006-01-02T15:04:05"), nil
}

func GetNextWeekDate(date string) (string, error) {
	if len(date) != 10 {
		return "", errors.New("err format")
	}

	newDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", err
	}

	return newDate.AddDate(0, 0, 7).Format("2006-01-02"), nil
}

func GetShortDisplayDate(date string) string {
	if strings.Index(date, "T") != -1 && len(date) > 10 {
		return date[:len(date)-9]
	} else {
		return date
	}
}

func GenSaltHash(pl string) string {
	hasher := sha256.New()
	io.WriteString(hasher, pl+Salt)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))[0:8]
}
