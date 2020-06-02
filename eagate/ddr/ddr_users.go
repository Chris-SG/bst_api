package ddr

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/ddr_models"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func PlayerInformationForClient(client util.EaClient) (playerDetails ddr_models.PlayerDetails, playcount ddr_models.Playcount, err bst_models.Error) {
	err = bst_models.ErrorOK
	document, err := playerInformationDocument(client)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	playerDetails, err = playerInformationFromPlayerDocument(document)
	if !err.Equals(bst_models.ErrorOK) {
		nameSelection := document.Find("div#dancer_name div.name_str").First()
		if nameSelection != nil && nameSelection.Text() == "---" {
			glog.Warningf("user %s has not played ddr", client.GetUserModel().Name)
			err = bst_models.ErrorDdrNotPlayed
		}
		return
	}
	playcount, err = playcountFromPlayerDocument(document)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	eaGateUser := client.GetUserModel().Name
	playerDetails.EaGateUser = &eaGateUser

	return
}

func playerInformationFromPlayerDocument(document *goquery.Document) (playerDetails ddr_models.PlayerDetails, err bst_models.Error) {
	err = bst_models.ErrorOK
	status := document.Find("table#status").First()
	if status == nil {
		err = bst_models.ErrorGormSelector
		return
	}
	statusDetails, err := util.TableThTd(status)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	playerDetails.Name = statusDetails["ダンサーネーム"]
	code, e := strconv.ParseInt(statusDetails["DDR-CODE"], 10, 32)
	if e != nil {
		err = bst_models.ErrorStringParse
		return
	}
	playerDetails.Code = int(code)
	playerDetails.Prefecture = statusDetails["所属都道府県"]
	playerDetails.SingleRank = statusDetails["段位(SINGLE)"]
	playerDetails.DoubleRank = statusDetails["段位(DOUBLE)"]
	playerDetails.Affiliation = statusDetails["所属クラス"]

	return
}

func playcountFromPlayerDocument(document *goquery.Document) (playcount ddr_models.Playcount, err bst_models.Error) {
	err = bst_models.ErrorOK
	status := document.Find("table#status").First()
	if status == nil {
		glog.Warningf("failed to find document field")
		err = bst_models.ErrorGormSelector
		return
	}
	single := document.Find("div#single table.small_table").First()
	if single == nil {
		glog.Warningf("failed to find document field")
		err = bst_models.ErrorGormSelector
		return
	}
	double := document.Find("div#double table.small_table").First()
	if double == nil {
		glog.Warningf("failed to find document field")
		err = bst_models.ErrorGormSelector
		return
	}

	numericalStripper, _ := regexp.Compile("[^0-9]+")
	timeFormat := "2006-01-02 15:04:05"
	timeLocation, e := time.LoadLocation("Asia/Tokyo")
	if e != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		err = bst_models.ErrorTimeLocLoad
		return
	}

	statusDetails, err := util.TableThTd(status)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Warningf("failed to find document field")
		return
	}
	singleDetails, err := util.TableThTd(single)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Warningf("failed to find document field")
		return
	}
	doubleDetails, err := util.TableThTd(double)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Warningf("failed to find document field")
		return
	}

	code, e := strconv.ParseInt(statusDetails["DDR-CODE"], 10, 32)
	if e != nil {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	playcount.PlayerCode = int(code)

	playcount.Playcount, e = strconv.Atoi(numericalStripper.ReplaceAllString(statusDetails["総プレー回数"], ""))
	if e != nil {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	playcount.LastPlayDate, e = time.ParseInLocation(timeFormat, statusDetails["最終プレー日時"], timeLocation)
	if e != nil {
		glog.Warningf("failed to parse time")
		err = bst_models.ErrorTimeParse
		return
	}

	playcount.SinglePlaycount, e = strconv.Atoi(numericalStripper.ReplaceAllString(singleDetails["プレー回数"], ""))
	if !err.Equals(bst_models.ErrorOK) {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	playcount.SingleLastPlayDate, e = time.ParseInLocation(timeFormat, singleDetails["最終プレー日時"], timeLocation)
	if e != nil {
		glog.Warningf("failed to parse time")
		err = bst_models.ErrorTimeParse
		return
	}

	playcount.DoublePlaycount, e = strconv.Atoi(numericalStripper.ReplaceAllString(doubleDetails["プレー回数"], ""))
	if e != nil {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	playcount.DoubleLastPlayDate, e = time.ParseInLocation(timeFormat, doubleDetails["最終プレー日時"], timeLocation)
	if e != nil {
		glog.Warningf("failed to parse time")
		err = bst_models.ErrorTimeParse
		return
	}

	return
}

func SongStatisticsForClient(client util.EaClient, charts []ddr_models.SongDifficulty, playerCode int) (songStatistics []ddr_models.SongStatistics, err bst_models.Error) {
	err = bst_models.ErrorOK
	mtx := &sync.Mutex{}

	wg := new(sync.WaitGroup)
	wg.Add(len(charts))

	errCount := 0

	for _, chart := range charts {
		go func (diff ddr_models.SongDifficulty) {
			defer wg.Done()
			document, err := musicDetailDifficultyDocument(client, diff.SongId, ddr_models.StringToMode(diff.Mode), ddr_models.StringToDifficulty(diff.Difficulty))
			if !err.Equals(bst_models.ErrorOK) {
				glog.Errorf("failed to load document for client %s: songid %s\n", client.GetUserModel().Name, diff.SongId)
				errCount++
				return
			}
			statistics, err := chartStatisticsFromDocument(document, playerCode, diff)
			if !err.Equals(bst_models.ErrorOK) {
				glog.Errorf("failed to load statistics for client %s: songid %s\n", client.GetUserModel().Name, diff.SongId)
				errCount++
				return
			}
			if statistics.PlayerCode == 0 {
				return
			}

			mtx.Lock()
			defer mtx.Unlock()
			songStatistics = append(songStatistics, statistics)
		}(chart)
	}

	wg.Wait()

	if errCount > 0 {
		glog.Warningf("failed loading all statistic for %s:  %d of %d errors\n", client.GetUserModel().Name, errCount, len(charts))
		err = bst_models.ErrorDdrStats
		return
	}

	glog.Infof("got %d statistics for user %s\n", len(songStatistics), client.GetUserModel().Name)
	return
}

func chartStatisticsFromDocument(document *goquery.Document, playerCode int, difficulty ddr_models.SongDifficulty) (songStatistics ddr_models.SongStatistics, err bst_models.Error) {
	err = bst_models.ErrorOK
	if strings.Contains(document.Find("div#popup_cnt").Text(), "NO PLAY") {
		glog.Warningf("failed to find substring")
		err = bst_models.ErrorStringSearch
		return
	}
	if strings.Contains(document.Find("div#popup_cnt").Text(), "難易度を選択してください。") {
		glog.Warningf("failed to find substring")
		err = bst_models.ErrorStringSearch
		return
	}

	statsTable := document.Find("table#music_detail_table").First()
	if statsTable == nil {
		glog.Warningf("cannot find music_detail_table")
		err = bst_models.ErrorGormSelector
		return
	}

	timeFormat := "2006-01-02 15:04:05"
	timeLocation, e := time.LoadLocation("Asia/Tokyo")
	if e != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		err = bst_models.ErrorTimeLocLoad
		return
	}

	details, err := util.TableThTd(statsTable)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	songStatistics.MaxCombo, e = strconv.Atoi(details["最大コンボ数"])
	if e != nil {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	songStatistics.ClearCount, e = strconv.Atoi(details["クリア回数"])
	if e != nil {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	songStatistics.PlayCount, e = strconv.Atoi(details["プレー回数"])
	if e != nil {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	songStatistics.BestScore, e = strconv.Atoi(details["ハイスコア"])
	if e != nil {
		glog.Warningf("failed to parse value")
		err = bst_models.ErrorStringParse
		return
	}
	songStatistics.Rank = details["ハイスコア時のダンスレベル"]
	songStatistics.Lamp = details["フルコンボ種別"]
	songStatistics.LastPlayed, e = time.ParseInLocation(timeFormat, details["最終プレー時間"], timeLocation)
	if e != nil {
		glog.Warningf("failed to parse time in location")
		err = bst_models.ErrorTimeParse
		return
	}

	songStatistics.SongId = difficulty.SongId
	songStatistics.Mode = difficulty.Mode
	songStatistics.Difficulty = difficulty.Difficulty

	songStatistics.PlayerCode = playerCode
	return
}

func RecentScoresForClient(client util.EaClient, playerCode int) (scores []ddr_models.Score, err bst_models.Error) {
	err = bst_models.ErrorOK
	document, err := recentScoresDocument(client)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	scores, err = recentScoresFromDocument(document, playerCode)
	return
}

// TODO: error handling
func recentScoresFromDocument(document *goquery.Document, playerCode int) (scores []ddr_models.Score, err bst_models.Error) {
	err = bst_models.ErrorOK
	timeFormat := "2006-01-02 15:04:05"
	timeLocation, e := time.LoadLocation("Asia/Tokyo")
	if e != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		err = bst_models.ErrorTimeLocLoad
		return
	}

	document.Find("table#data_tbl tbody tr").Each(func(i int, s *goquery.Selection) {
		if s.Find("td").Length() == 0 {
			return
		}

		score := ddr_models.Score{}

		info := s.Find("a.music_info.cboxelement").First()
		href, exists := info.Attr("href")
		if !exists {
			return
		}
		difficulty, e := strconv.Atoi(href[len(href)-1:])
		if e != nil {
			glog.Warningf("failed to parse value")
			err = bst_models.ErrorStringParse
			return
		}

		score.Mode = ddr_models.Mode(difficulty / 5).String()
		if ddr_models.StringToMode(score.Mode) == ddr_models.Double {
			difficulty++
		}
		score.Difficulty = ddr_models.Difficulty(difficulty % 5).String()
		score.SongId = href[strings.Index(href, "=")+1 : strings.Index(href, "&")]

		score.Score, _ = strconv.Atoi(s.Find("td.score").First().Text())

		timeSelection := s.Find("td.date").First()
		t, e := time.ParseInLocation(timeFormat, timeSelection.Text(), timeLocation)
		if e != nil {
			glog.Warningf("failed to parse time in location")
			err = bst_models.ErrorTimeParse
			return
		}
		score.TimePlayed = t

		rankSelection := s.Find("td.rank").First()
		imgSelection := rankSelection.Find("img").First()
		path, exists := imgSelection.Attr("src")
		if exists {
			score.ClearStatus = !strings.Contains(path, "rank_s_e")
		}

		score.PlayerCode = playerCode

		scores = append(scores, score)
	})

	return
}

func WorkoutDataForClient(client util.EaClient, playerCode int) (workoutData []ddr_models.WorkoutData, err bst_models.Error) {
	err = bst_models.ErrorOK
	document, err := workoutDocument(client)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	workoutData, err = workoutDataFromDocument(document, playerCode)
	return
}

func workoutDataFromDocument(document *goquery.Document, playerCode int) (workoutData []ddr_models.WorkoutData, err bst_models.Error) {
	err = bst_models.ErrorOK
	format := "2006-01-02"
	loc, e := time.LoadLocation("Asia/Tokyo")
	if e != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		err = bst_models.ErrorTimeLocLoad
		return
	}

	table := document.Find("table#work_out_left")
	if table.Length() == 0 {
		glog.Warningf("failed to find field in document")
		err = bst_models.ErrorGormSelector
		return
	}

	tableBody := table.First().Find("tbody").First()
	if tableBody == nil {
		glog.Warningf("failed to find field in document")
		err = bst_models.ErrorGormSelector
		return
	}

	tableBody.Find("tr").Each(func(i int, s *goquery.Selection) {
		if s.Find("td").Length() == 5 {
			wd := ddr_models.WorkoutData{}
			s.Find("td").Each(func(i int, dataSelection *goquery.Selection) {
				if i == 1 {
					t, e := time.ParseInLocation(format, dataSelection.Text(), loc)
					if e == nil {
						wd.Date = t
					}
				} else if i == 2 {
					numerical, e := regexp.Compile("[^0-9]+")
					if e == nil {
						panic(e)
					}
					numericStr := numerical.ReplaceAllString(dataSelection.Text(), "")
					wd.PlayCount, _ = strconv.Atoi(numericStr)
				} else if i == 3 {
					numerical, e := regexp.Compile("[^0-9.]+")
					if e != nil {
						glog.Errorf("regex failure! %s\n", e.Error())
						panic(e)
					}
					numericStr := numerical.ReplaceAllString(dataSelection.Text(), "")
					kcalFloat, err := strconv.ParseFloat(numericStr, 32)
					if err == nil {
						wd.Kcal = float32(kcalFloat)
					}
				}
			})
			wd.PlayerCode = playerCode
			workoutData = append(workoutData, wd)
		}
	})
	return
}
