package main

/*
import (
	"bytes"
	"context"
	"github.com/bitly/go-simplejson"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func QueryPtByUrl(paipu_url string) (int, error) {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:ygtianlou@localhost:5432/pgtest")
	if err != nil {
		log.Println("db connect fail", err)
		return 0, err
	}
	defer conn.Close(context.Background())

	var cnt int
	sqlQuery := `select count(*) as num from pt where paipu_url = $1`
	err = conn.QueryRow(context.Background(), sqlQuery, paipu_url).Scan(&cnt)
	if err != nil {
		log.Println("query fail", err)
		return 0, err
	}

	return cnt, nil
}

func main() {

	skipPaipu := []string{
		"231112-a79ea4d3-ecfa-4dde-95d9-41f90e891bda",
		"231112-e68860bf-8180-41c1-8bc8-e3170b22d82e",
		"231112-50750052-3ae5-4b6a-827c-84bce17bc7bb",
		"231112-0133df33-4e9b-4cef-8e0e-7f2591eefbe7",
		"231112-34bce479-e8a4-43df-92fa-640242b7243a",
		"231112-0c583b1b-78df-4ab5-aa03-87719e7bef42",
		"231112-45ef7472-e7d3-4ff9-98e2-be4f38040954",
		"231112-87f9c5cd-a4e6-46aa-933f-2137e6dc7a0a",
		"231112-51257c67-1f2f-431e-bc78-eb04c1c5c788",
		"231112-85738420-ccb2-4490-916d-b6ce1515b8ca",
	}
	skipTo := "paipu=231124-dabcf998-24ac-4cd4-9940-5b34d2cb1a3b"
	bolSkip := true
	bolUseSkip := false

	// read file
	filePaipu := "./test/script/log_paipu_2401.txt"
	//filePaipu := "./test/script/log_paipu.20231207.txt"
	b, err := os.ReadFile(filePaipu)
	if err != nil {
		log.Fatal(err)
	}

	txt := string(b)

	// regex match
	//txt := `{"comboRecords":[{"url":"雀魂牌谱:https://game.maj-soul.com/1/?paipu=231112-3d310aa6-4d19-44a4-b21a-d6f81192dc61_a31266999"},{"url":"雀魂牌谱:https://game.maj-soul.com/1/?paipu=231112-59664962-34b5-46df-a25e-9071e6fe6335_a31266999"},{"url":"雀魂牌谱:https://game.maj-soul.com/1/?paipu=231112-0085cc92-43aa-47b4-82a5-7be688836f3a_a31266999"},{"url":""}],"mode":"53"}`
	pattern := `paipu=[0-9]+\-[0-9a-z]+\-[0-9a-z]+\-[0-9a-z]+\-[0-9a-z]+\-[0-9a-z]+_[0-9a-z]+`
	reg := regexp.MustCompile(pattern)
	found := reg.FindAllString(txt, -1)
	if len(found) == 0 {
		log.Fatal("no found")
	}

	log.Println("regex found", len(found))

	for _, oneLine := range found {

		oneLastIdx := strings.LastIndex(oneLine, "_")
		if oneLastIdx == -1 {
			log.Println("paipu missing _", oneLine)
			continue
		}
		oneLine = oneLine[:oneLastIdx]

		if bolUseSkip {
			if oneLine == skipTo {
				bolSkip = false
			}

			if bolSkip {
				continue
			}
		}

		oneCnt, e2 := QueryPtByUrl(oneLine[6:])
		if e2 != nil {
			log.Fatal(e2)
		}

		if bolUseSkip {
			for j := 0; j < len(skipPaipu); j++ {
				if skipPaipu[j] == oneLine[6:] {
					oneCnt = 1
					break
				}
			}
		}

		if oneCnt > 0 {
			//log.Println("skip", oneLine)
			continue
		}

		//log.Println(oneLine)
		oneMap := make(map[string]string)
		oneMap["url"] = oneLine + "_"
		var arrMap []map[string]string
		arrMap = append(arrMap, oneMap)

		js := simplejson.New()
		js.Set("comboRecords", arrMap)
		oneLog, err := js.Encode()
		//log.Println(string(oneLog), err)

		// http add
		r, err := http.Post("http://127.0.0.1/api/addMajsoulRecord", "application/json", bytes.NewBuffer(oneLog))
		if err != nil {
			log.Fatal(err)
		}

		if r.StatusCode != 200 {
			log.Fatalln("fail", oneLine)
		} else {
			log.Println("ok", oneLine)
		}

		time.Sleep(time.Second / 2)
	}
}
*/
