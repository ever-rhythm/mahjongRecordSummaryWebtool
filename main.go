package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/mahjongRecordSummaryWebtool/action"
	"github.com/mahjongRecordSummaryWebtool/client"
	"github.com/mahjongRecordSummaryWebtool/server"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	// startup server or cron
	flagMode := flag.String("mode", "server", "server of cron")
	flag.Parse()

	// timezone
	err := os.Setenv("TZ", "Asia/Shanghai")
	if err != nil {
		log.Fatalln(err)
	}

	if *flagMode == "server" {
		// init log
		logFile, err := os.Create("./gin.log." + time.Now().Format("2006-01-02.15-04-05"))
		if err != nil {
			log.Fatalln(err)
		}
		log.SetOutput(logFile)

		// gin
		gin.DisableConsoleColor()
		gin.DefaultWriter = io.MultiWriter(logFile)
		gin.SetMode(gin.DebugMode)
		r := gin.Default()

		// load config from file
		err = server.LoadConfigFile()
		if err != nil {
			log.Fatalln("load config.json fail , exit", err)
		}

		// init majsoul config from http
		err = server.InitMajsoulConfig()
		if err != nil {
			log.Fatalln("InitMajsoulConfig fail , exit", client.MajsoulServerConfig, err)
		}

		// cron
		//go server.CronUpdateMajsoulConfig()
		go server.CronUpdatePlayerCareer()

		// static
		r.StaticFS("/static", http.Dir("./web/static"))
		r.LoadHTMLFiles(
			"web/index.html",
			"web/rank.html",
			"web/trend.html",
			"web/debug.html",
			"web/vs.html",
			"web/about.html",
			"web/nav.html",
			"web/plcr.html",
			"web/adc.html",
		)

		// get
		r.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", "home"); return })
		r.GET("/index.html", func(c *gin.Context) { c.HTML(200, "index.html", "index"); return })
		r.GET("/debug.html", func(c *gin.Context) { c.HTML(200, "debug.html", "debug"); return })
		r.GET("/rank.html", func(c *gin.Context) { c.HTML(200, "rank.html", "rank"); return })
		r.GET("/trend.html", func(c *gin.Context) { c.HTML(200, "trend.html", "trend"); return })
		r.GET("/vs.html", func(c *gin.Context) { c.HTML(200, "vs.html", "vs"); return })
		r.GET("/nav.html", func(c *gin.Context) { c.HTML(200, "nav.html", "nav"); return })
		r.GET("/about.html", func(c *gin.Context) { c.HTML(200, "about.html", "about"); return })
		r.GET("/plcr.html", func(c *gin.Context) { c.HTML(200, "plcr.html", "plcr"); return })
		r.GET("/adc.html", func(c *gin.Context) { c.HTML(200, "adc.html", "adc"); return })

		// post
		r.POST("/api/v2", action.Summary)
		r.POST("/api/group_rank", action.GroupRank)
		r.POST("/api/group_player_trend", action.GroupPlayerTrend)
		r.POST("/api/group_player_op_trend", action.GroupPlayerOpponentTrend)
		r.POST("/api/competitor", action.Competitor)
		r.POST("/api/nav", action.GetNavJson)
		r.POST("/api/plcr", action.PlayerCareer)
		r.POST("/api/adc_summary", action.AdcSummary)
		r.POST("/api/add_adc_vip", action.AddPlayerVip)

		// server
		err = r.Run(":" + server.ConfigGin.Port)
		if err != nil {
			log.Fatalln("gin run server fail, exit", err)
		}

	} else if *flagMode == "cron" {
		/*
			// init log
			logFile, err := os.Create("./cron.gin.log." + time.Now().Format("2006-01-02.15-04-05"))
			if err != nil {
				log.Fatalln(err)
			}
			log.SetOutput(logFile)

			// load config from file
			err = server.LoadConfigFile()
			if err != nil {
				log.Fatalln("load config.json fail , exit", err)
			}

			// cron
			go server.CronUpdatePlayerCareer()

		*/
	} else {
		log.Fatalln("not support mode", *flagMode)
	}
}
