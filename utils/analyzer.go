package utils

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// test done
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

// test done
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
		num += 1
		if num > 7 {
			num = 1
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
			return nil, errors.New("invalid record url " + oneUrl)
		}

		oneLastIdx := strings.LastIndex(oneUrl, "_")

		arrUuid = append(arrUuid, oneUrl[oneIdx+len("paipu="):oneLastIdx])
	}

	if len(arrUuid) == 0 {
		return arrUuid, errors.New("empty url")
	}

	return arrUuid, nil
}

func GetRateZhuyiByMode(mode string) (int, int, error) {
	type stMode struct {
		rate  int
		zhuyi int
	}

	mapSupportMode := map[string]stMode{}
	mapSupportMode["1"] = stMode{1, 0}
	mapSupportMode["2"] = stMode{2, 0}
	mapSupportMode["5"] = stMode{5, 0}
	mapSupportMode["10"] = stMode{10, 0}
	mapSupportMode["20"] = stMode{20, 0}

	mapSupportMode["13"] = stMode{1, 3}
	mapSupportMode["23"] = stMode{2, 3}
	mapSupportMode["53"] = stMode{5, 3}
	mapSupportMode["103"] = stMode{10, 3}
	mapSupportMode["203"] = stMode{20, 3}

	v, b := mapSupportMode[mode]

	if !b {
		return 0, 0, errors.New("mode not support")
	} else {
		return v.rate, v.zhuyi, nil
	}
}

func GetMajsoulBot() (string, string) {
	type stBot struct {
		n string
		p string
	}

	mapBot := map[int]stBot{}
	mapBot[0] = stBot{"3264373548@qq.com", "77shuafen"}
	mapBot[1] = stBot{"evershikieiki@gmail.com", "77shuafen"}
	mapBot[2] = stBot{"343669220@qq.com", "77shuafen"}

	idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(mapBot))

	return mapBot[idx].n, mapBot[idx].p
}

func GetMajSoulBotByIdx(idx int) (string, string) {
	account := []string{
		"3264373548@qq.com",
		"evershikieiki@gmail.com",
		"343669220@qq.com",
		"everhythm@aliyun.com",
		"everhythm@sina.cn",
		"everhythm@sohu.com",
	}
	p := []string{
		"77shuafen",
		"77shuafen",
		"77shuafen",
		"887178",
		"77shuafen",
		"77shuafen",
	}

	if idx < len(p) {
		return account[idx], p[idx]
	} else {
		return "", ""
	}
}
