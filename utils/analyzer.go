package utils

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
)

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
	} else if tp == "m" || tp == "p" || tp == "s"{
		num += 1
		if num > 9 {
			num = 1
		}
	} else {
		return "", errors.New("invalid hand char")
	}

	return strconv.Itoa(num) + tp, nil
}

func GetUuidByRecordUrl(arrUrl []string) ([]string, error){
	var arrUuid []string

	for _,oneUrl := range arrUrl {
		oneIdx := strings.Index(oneUrl, "paipu=")
		if oneIdx == -1 {
			return nil, errors.New("invalid record url " + oneUrl)
		}

		oneLastIdx := strings.LastIndex(oneUrl, "_")

		arrUuid = append(arrUuid, oneUrl[oneIdx+len("paipu="):oneLastIdx])
	}

	return arrUuid, nil
}

func GetRateZhuyiByMode(mode string) (int, int, error) {
	type stMode struct {
		rate int
		zhuyi int
	}

	mapSupportMode := map[string]stMode{}
	mapSupportMode["2"] = stMode{2,0}
	mapSupportMode["5"] = stMode{5,0}
	mapSupportMode["10"] = stMode{10,0}
	mapSupportMode["20"] = stMode{20,0}
	mapSupportMode["23"] = stMode{2,3}
	mapSupportMode["53"] = stMode{5,3}
	mapSupportMode["103"] = stMode{10,3}
	mapSupportMode["203"] = stMode{20,3}

	v,b := mapSupportMode[mode]

	if ! b {
		return 0,0, errors.New("mode not support")
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
	mapBot[0] = stBot{"3264373548@qq.com","77shuafen"}
	mapBot[1] = stBot{"evershikieiki@gmail.com","77shuafen"}
	mapBot[2] = stBot{"343669220@qq.com","77shuafen"}

	idx := rand.Intn(len(mapBot))

	return mapBot[idx].n, mapBot[idx].p
}