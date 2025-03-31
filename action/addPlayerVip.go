package action

import (
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/api"
	"github.com/mahjongRecordSummaryWebtool/server"
	"log"
)

func AddPlayerVip(c *gin.Context) {
	objJs := okRet()

	b, _ := c.GetRawData()
	req, err := simplejson.NewJson(b)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "请求格式有误")
		c.JSON(500, objJs)
		return
	}

	pl := req.Get("pl").MustString()
	if len(pl) == 0 {
		objJs.Set("msg", "请求参数有误")
		c.JSON(500, objJs)
		return
	}

	retCode, err := api.AddVip(pl)
	if err != nil {
		log.Println(err)
		objJs.Set("msg", "添加失败")
		c.JSON(500, objJs)
		return
	}

	// update config cache
	server.UpdateVipConfig()

	objJs.SetPath([]string{"data", "code"}, retCode)
	c.JSON(200, objJs.MustMap())
	return
}
