package test

// logout
/*
	defer func() {
		//mSoul.LobbyConn.Close()

		reqLogout := message.ReqLogout{}
		rspLogout, err := mSoul.Logout(mSoul.Ctx, &reqLogout)
		if err != nil {
			log.Println(err)
		}

		if rspLogout.Error != nil {
			log.Println(rspLogout.Error)
		}
	}()

*/

// get record socket read error ?
/*
	type result struct {
		s  string
		bs []byte
	}
	var mapUuidBytes = make(map[string][]byte)
	var chanResult = make(chan result, 12)
	defer close(chanResult)

	for _, oneUuid := range uuids {
		go func(uuid string) {
			reqPaipu := message.ReqGameRecord{
				GameUuid:            uuid,
				ClientVersionString: mSoul.Version.Web(),
			}
			resPaipu, err := mSoul.FetchGameRecord(mSoul.Ctx, &reqPaipu)
			if err != nil {
				log.Println("paipu fail", err)
				chanResult <- result{
					s:  uuid,
					bs: nil,
				}
			} else {
				chanResult <- result{
					s:  uuid,
					bs: resPaipu.Data,
				}
			}
		}(oneUuid)
	}

	for i := 0; i < len(uuids); i++ {
		oneResult := <-chanResult
		if len(oneResult.bs) == 0 {
			log.Println("paipu empty" + oneResult.s)
			return nil, nil, nil, errors.New("paipu fail")
		}
		mapUuidBytes[oneResult.s] = oneResult.bs
	}

*/

// logout 好像没啥用 defer conn loop panic
/*
	reqLogout := message.ReqLogout{}
	rspLogout, err := mSoul.Logout(mSoul.Ctx, &reqLogout)
	if err != nil {
		log.Println(err)
	}
	if rspLogout.Error != nil {
		log.Println(rspLogout.Error)
	}
*/
