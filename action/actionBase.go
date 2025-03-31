package action

import "github.com/bitly/go-simplejson"

func okRet() *simplejson.Json {
	objSucc := simplejson.New()
	objSucc.Set("status", 0)
	objSucc.Set("msg", "成功")
	return objSucc
}
