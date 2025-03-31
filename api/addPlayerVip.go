package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"time"
)

func AddVip(pl string) (string, error) {
	ret, err := utils.QueryVip(pl)
	if err != nil {
		log.Println(err)
		return "", errors.New("内部错误")
	}

	if len(ret) > 0 {
		return "", errors.New("用户名已登记")
	}

	code := utils.GenSaltHash(pl)
	timeNow := time.Now()
	timeIndex := time.Date(timeNow.Year(), timeNow.Month(), 1, 0, 0, 0, 0, time.Local)
	timeEnd := timeIndex.AddDate(1, 0, 0)

	err = utils.InsertVipAndCareer(pl, code, timeIndex, timeNow, timeEnd)
	if err != nil {
		log.Println(err)
		return "", errors.New("创建失败")
	}

	return code, nil
}
