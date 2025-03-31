package server

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/utils"
	"log"
	"math/rand"
	"os"
	"time"
)

type configGin struct {
	Domain string
	Port   string
}

type configCron struct {
	TimestampCronUpdatePlayerCareer int64
	IntervalCronUpdatePlayerCareer  int64
}

var ConfigGin = configGin{}

var ConfigCron = configCron{
	TimestampCronUpdatePlayerCareer: 0,
	IntervalCronUpdatePlayerCareer:  30,
}

// LoadConfigFile load config json file
func LoadConfigFile() error {

	btConfig, err := os.ReadFile("./config.json")
	if err != nil {
		log.Println("load config.json fail")
		return err
	}

	objJsConfig, err := simplejson.NewJson(btConfig)
	if err != nil {
		log.Println("conv config.json to json fail")
		return err
	}

	// set config
	ConfigGin.Domain = objJsConfig.Get("server").Get("domain").MustString()
	ConfigGin.Port = objJsConfig.Get("server").Get("port").MustString()

	client.MajsoulServerConfig.Host = objJsConfig.Get("MajsoulServerConfig").Get("host").MustString()
	client.MajsoulServerConfig.Ws_server = objJsConfig.Get("MajsoulServerConfig").Get("ws_server").MustString()
	client.MajsoulServerConfig.CodeUpdateInterval = objJsConfig.Get("MajsoulServerConfig").Get("CodeUpdateInterval").MustInt64()

	utils.ConfigDb.Dsn = objJsConfig.Get("postgre").Get("dsn").MustString()
	utils.ConfigDb.TimeoutConnect = time.Duration(objJsConfig.Get("postgre").Get("timeout").MustInt64())

	utils.ConfigMajsoulBot.Acc = objJsConfig.Get("MajsoulBot").Get("acc").MustStringArray()
	utils.ConfigMajsoulBot.Pwd = objJsConfig.Get("MajsoulBot").Get("pwd").MustStringArray()

	utils.ConfigMode.Mode = objJsConfig.Get("modeMap").Get("mode").MustStringArray()
	utils.ConfigMode.Rate = objJsConfig.Get("modeMap").Get("rate").MustStringArray()
	utils.ConfigMode.Zy = objJsConfig.Get("modeMap").Get("zy").MustStringArray()
	utils.ConfigMode.RecordContestIds = objJsConfig.Get("recordContestUid").MustStringArray()
	utils.ConfigMode.RecordMode = objJsConfig.Get("recordMode").MustInt()
	utils.ConfigMode.RecordModeList = objJsConfig.Get("recordModeList").MustStringArray()
	utils.ConfigMode.RecordSwitch = objJsConfig.Get("RecordSwitch").MustInt()
	utils.ConfigMode.MapRecordVips = make(map[string]int)

	ConfigCron.IntervalCronUpdatePlayerCareer = objJsConfig.Get("cron").Get("IntervalCronUpdatePlayerCareer").MustInt64()

	// update cache from db
	UpdateVipConfig()

	// validate config
	if len(utils.ConfigMajsoulBot.Acc) != len(utils.ConfigMajsoulBot.Pwd) ||
		len(utils.ConfigMajsoulBot.Acc) == 0 {
		log.Println("config err, bot acc pwd len not match")
		return errors.New("bot err")
	}

	if len(utils.ConfigMode.Mode) != len(utils.ConfigMode.Rate) ||
		len(utils.ConfigMode.Mode) != len(utils.ConfigMode.Zy) ||
		len(utils.ConfigMode.Mode) == 0 {
		log.Println("config err, mode rate zy len not match")
		return errors.New("mode err")
	}

	log.Println("load config.json ok")

	return nil
}

func InitMajsoulConfig() error {

	// http client
	request := utils.NewRequest(client.MajsoulServerConfig.Host)
	randv := int(rand.Float32()*1000000000) + int(rand.Float32()*1000000000)

	// get version
	rspVersion, err := request.Get(fmt.Sprintf("1/version.json?randv=%d", randv))
	if err != nil {
		log.Println("InitMajsoulConfig get version host fail", client.MajsoulServerConfig, err)
		return err
	}

	objJsVersion, err := simplejson.NewJson(rspVersion)
	if err != nil {
		log.Println("InitMajsoulConfig unpack version fail", err)
		return err
	}

	// update cache timestamp
	client.MajsoulServerConfig.CodeUpdateTimestamp = time.Now().Unix()
	client.MajsoulServerConfig.Version, err = objJsVersion.Get("version").String()
	client.MajsoulServerConfig.Force_version, err = objJsVersion.Get("force_version").String()
	client.MajsoulServerConfig.Code, err = objJsVersion.Get("code").String()

	log.Println("InitMajsoulConfig ok", client.MajsoulServerConfig)

	return nil
}

// UpdateVipConfig update config cache
func UpdateVipConfig() {
	retVip, err := utils.QueryVips()
	if err != nil {
		log.Println("updateVipConfig fail", err)
		return
	}

	for _, v := range retVip {
		utils.ConfigMode.MapRecordVips[v.Name] = 1
	}

	log.Println("updateVipConfig ok", utils.ConfigMode.MapRecordVips)
}
