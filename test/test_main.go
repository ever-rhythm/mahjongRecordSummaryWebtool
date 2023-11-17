package test

/*
	// get region
	rspRegion, err := request.Get("1/v" + client.MajsoulServerConfig.Version + "/config.json")
	if err != nil {
		log.Println("get region fail", client.MajsoulServerConfig, err)
		return err
	}

	objJsRegion, err := simplejson.NewJson(rspRegion)
	if err != nil {
		log.Println("unpack region fail", err)
		return err
	}

	client.MajsoulServerConfig.Region_urls[0], err = objJsRegion.Get("ip").GetIndex(0).Get("region_urls").GetIndex(0).Get("url").String()
	if err != nil {
		log.Println("unpack region json fail", err)
		return err
	}

*/

// get ws
/*
	requestLb := utils.NewRequest(client.MajsoulServerConfig.Region_urls[0])
	rspWs, err := requestLb.Get(fmt.Sprintf("?service=ws-gateway&protocol=ws&ssl=true&rv=%d", randv))
	if err != nil {
		log.Println("get ws fail", client.MajsoulServerConfig, err)
		return err
	}

	objJsWs, err := simplejson.NewJson(rspWs)
	if err != nil {
		log.Println("unpack ws fail", err)
		return err
	}

	tmpWs, err := objJsWs.Get("servers").GetIndex(0).String()
	client.MajsoulServerConfig.Ws_server = "wss://" + tmpWs + "/gateway"
	if err != nil {
		log.Println("unpack ws json fail", err)
		return err
	}

*/

// custom ws
//client.MajsoulServerConfig.Ws_server = "wss://gateway-sy.maj-soul.com:443/gateway"
//client.MajsoulServerConfig.Ws_server = "wss://gateway-hw.maj-soul.com:443/gateway"

/*
	var rows []map[string]string

	var row = make(map[string]string)
	row["desc"] = "总计点数"
	row["p0"] = "5"
	row["p1"] = "15"
	row["p2"] = "-5"
	row["p3"] = "-15"
	rows = append(rows, row)

	row = make(map[string]string)
	row["desc"] = "总计祝仪"
	row["p0"] = "1"
	row["p1"] = "0"
	row["p2"] = "-3"
	row["p3"] = "2"
	rows = append(rows, row)

	row = make(map[string]string)
	row["desc"] = "最终积分"
	row["p0"] = "100"
	row["p1"] = "0"
	row["p2"] = "-300"
	row["p3"] = "200"
	rows = append(rows, row)

	js := simplejson.New()
	js.Set("status",0)
	js.Set("msg","ok")
	js.SetPath([]string{"data","names"}, []string{"雀士1","雀士2","雀士3","雀士4"})
	js.SetPath([]string{"data","rows"}, rows)
	js.SetPath([]string{"data","ptRows"}, rows)
	js.SetPath([]string{"data","zhuyiRows"}, rows)

	tmpMap, err := js.Map()
	if err != nil{
		fmt.Println(err)
	}

	//fmt.Println(tmpMap)
	c.JSON(200, tmpMap)

*/

/*
	r.POST("/api/test", func(c *gin.Context) {
		b, _ := c.GetRawData()
		fmt.Println(string(b))
		req, err := simplejson.NewJson(b)
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}

		var records []string
		mapRecords, err := req.Get("comboRecords").Array()
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}
		for _, v := range mapRecords {
			tmpMap, _ := v.(map[string]interface{})
			tmpRecord, _ := tmpMap["url"].(string)
			records = append(records, tmpRecord)
		}

		mode, err := req.Get("mode").String()
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}

		//fmt.Println(records, mode)
		mapPlayerInfo, ptRows, zhuyiRows, err := api.GetSummaryByRecords(records, mode)
		if err != nil {
			c.JSON(500, err)
		}

		// names
		var arrName []string
		for i := 0; i < 4; i++ {
			for _, oneInfo := range mapPlayerInfo {
				if i == int(oneInfo.Seat) {
					arrName = append(arrName, oneInfo.Nickname)
					break
				}
			}
		}

		var rows []map[string]string
		var rowPtSum = make(map[string]string)
		var rowZhuyiSum = make(map[string]string)
		var rowMoneySum = make(map[string]string)

		rowPtSum["desc"] = "总计点数"
		rowZhuyiSum["desc"] = "总计祝仪"
		rowMoneySum["desc"] = "最终积分"

		for _, oneInfo := range mapPlayerInfo {
			oneIdx := "p" + strconv.Itoa(int(oneInfo.Seat))
			rowPtSum[oneIdx] = fmt.Sprintf("%.1f", float32(oneInfo.TotalPoint)/1000)
			rowZhuyiSum[oneIdx] = strconv.Itoa(int(oneInfo.Zhuyi))
			rowMoneySum[oneIdx] = oneInfo.Sum
		}

		rows = append(rows, rowPtSum, rowZhuyiSum, rowMoneySum)

		objJs := simplejson.New()
		objJs.Set("status", 0)
		objJs.Set("msg", "ok")
		objJs.SetPath([]string{"data", "names"}, arrName)
		objJs.SetPath([]string{"data", "rows"}, rows)
		objJs.SetPath([]string{"data", "ptRows"}, ptRows)
		objJs.SetPath([]string{"data", "zhuyiRows"}, zhuyiRows)

		tmpMap, err := objJs.Map()
		if err != nil {
			fmt.Println(err)
			c.JSON(500, "")
		}

		//fmt.Println(tmpMap)
		c.JSON(200, tmpMap)

	})
*/
/*

	client.MajsoulServerConfig.Version = "0.10.292.w"
	client.MajsoulServerConfig.Force_version = "0.10.0.w"
	client.MajsoulServerConfig.Code = "v0.10.292.w/code.js"
	client.MajsoulServerConfig.Ws_server = "wss://gateway-sy.maj-soul.com:443/gateway"
*/
