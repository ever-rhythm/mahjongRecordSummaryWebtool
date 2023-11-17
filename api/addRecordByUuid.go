package api

import (
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"time"
)

// todo conv url to struct
func addRecordByUuid(uuid string) {

	var onePt = utils.Table_pt{
		Id:           0,
		Paipu_url:    "",
		Player_count: 0,
		State:        0,
		Desc:         "",
		Timestamp:    time.Time{},
		Pl_1:         "",
		Pl_2:         "",
		Pl_3:         "",
		Pl_4:         "",
		Pt_1:         0,
		Pt_2:         0,
		Pt_3:         0,
		Pt_4:         0,
		Zy_1:         0,
		Zy_2:         0,
		Zy_3:         0,
		Zy_4:         0,
	}

	_, err := utils.InsertPt(onePt)
	if err != nil {
		log.Println("insert fail", err)
	}
}
