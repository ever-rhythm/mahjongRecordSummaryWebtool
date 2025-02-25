package api

import (
	"errors"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
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
	err = utils.InsertVipAndCareer(pl, code)
	if err != nil {
		log.Println(err)
		return "", errors.New("创建失败")
	}

	return code, nil
}
