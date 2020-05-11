package ddr

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/golang/glog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func PlayerInformationForClient(client util.EaClient) (playerDetails ddr_models.PlayerDetails, playcount ddr_models.Playcount, err error) {
	document, err := playerInformationDocument(client)
	if err != nil {
		return
	}
	playerDetails, err = playerInformationFromPlayerDocument(document)
	if err != nil {
		return
	}
	playcount, err = playcountFromPlayerDocument(document)
	if err != nil {
		return
	}
	eaGateUser := client.GetUsername()
	playerDetails.EaGateUser = &eaGateUser

	return
}

func playerInformationFromPlayerDocument(document *goquery.Document) (playerDetails ddr_models.PlayerDetails, err error) {
	status := document.Find("table#status").First()
	if status == nil {
		err = fmt.Errorf("cannot find status table")
		return
	}
	statusDetails, err := util.TableThTd(status)
	if err != nil {
		return
	}
	playerDetails.Name = statusDetails["ダンサーネーム"]
	code, err := strconv.ParseInt(statusDetails["DDR-CODE"], 10, 32)
	if err != nil {
		return
	}
	playerDetails.Code = int(code)
	playerDetails.Prefecture = statusDetails["所属都道府県"]
	playerDetails.SingleRank = statusDetails["段位(SINGLE)"]
	playerDetails.DoubleRank = statusDetails["段位(DOUBLE)"]
	playerDetails.Affiliation = statusDetails["所属クラス"]

	return
}

func playcountFromPlayerDocument(document *goquery.Document) (playcount ddr_models.Playcount, err error) {
	status := document.Find("table#status").First()
	if status == nil {
		err = fmt.Errorf("cannot find status table")
		return
	}
	single := document.Find("div#single table.small_table").First()
	if single == nil {
		err = fmt.Errorf("cannot find single table")
		return
	}
	double := document.Find("div#double table.small_table").First()
	if double == nil {
		err = fmt.Errorf("cannot find double table")
		return
	}

	numericalStripper, _ := regexp.Compile("[^0-9]+")
	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	statusDetails, err := util.TableThTd(status)
	if err != nil {
		return
	}
	singleDetails, err := util.TableThTd(single)
	if err != nil {
		return
	}
	doubleDetails, err := util.TableThTd(double)
	if err != nil {
		return
	}

	code, err := strconv.ParseInt(statusDetails["DDR-CODE"], 10, 32)
	if err != nil {
		return
	}
	playcount.PlayerCode = int(code)

	playcount.Playcount, err = strconv.Atoi(numericalStripper.ReplaceAllString(statusDetails["総プレー回数"], ""))
	if err != nil {
		return
	}
	playcount.LastPlayDate, err = time.ParseInLocation(timeFormat, statusDetails["最終プレー日時"], timeLocation)
	if err != nil {
		return
	}

	playcount.SinglePlaycount, err = strconv.Atoi(numericalStripper.ReplaceAllString(singleDetails["プレー回数"], ""))
	if err != nil {
		return
	}
	playcount.SingleLastPlayDate, err = time.ParseInLocation(timeFormat, singleDetails["最終プレー日時"], timeLocation)
	if err != nil {
		return
	}

	playcount.DoublePlaycount, err = strconv.Atoi(numericalStripper.ReplaceAllString(doubleDetails["プレー回数"], ""))
	if err != nil {
		return
	}
	playcount.DoubleLastPlayDate, err = time.ParseInLocation(timeFormat, doubleDetails["最終プレー日時"], timeLocation)
	if err != nil {
		return
	}

	return
}

func SongStatisticsForClient(client util.EaClient, charts []ddr_models.SongDifficulty, playerCode int) (songStatistics []ddr_models.SongStatistics, err error) {
	mtx := &sync.Mutex{}

	wg := new(sync.WaitGroup)
	wg.Add(len(charts))

	errCount := 0

	for _, chart := range charts {
		go func (diff ddr_models.SongDifficulty) {
			defer wg.Done()
			document, err := musicDetailDifficultyDocument(client, diff.SongId, ddr_models.StringToMode(diff.Mode), ddr_models.StringToDifficulty(diff.Difficulty))
			if err != nil {
				glog.Errorf("failed to load document for client %s: songid %s\n", client.GetUsername(), diff.SongId)
				errCount++
				return
			}
			statistics, err := chartStatisticsFromDocument(document, playerCode, diff)
			if err != nil {
				glog.Errorf("failed to load statistics for client %s: songid %s\n", client.GetUsername(), diff.SongId)
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
		glog.Warningf("failed loading all statistic for %s:  %d of %d errors\n", client.GetUsername(), errCount, len(charts))
		err = fmt.Errorf("failed to load %d of %d chart statistics", errCount, len(charts))
		return
	}

	glog.Infof("got %d statistics for user %s\n", len(songStatistics), client.GetUsername())
	return
}

func chartStatisticsFromDocument(document *goquery.Document, playerCode int, difficulty ddr_models.SongDifficulty) (songStatistics ddr_models.SongStatistics, err error) {
	if strings.Contains(document.Find("div#popup_cnt").Text(), "NO PLAY") {
		return
	}
	if strings.Contains(document.Find("div#popup_cnt").Text(), "難易度を選択してください。") {
		return
	}

	statsTable := document.Find("table#music_detail_table").First()
	if statsTable == nil {
		err = fmt.Errorf("cannot find music_detail_table")
		return
	}

	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	details, err := util.TableThTd(statsTable)
	if err != nil {
		return
	}
	songStatistics.MaxCombo, err = strconv.Atoi(details["最大コンボ数"])
	songStatistics.ClearCount, err = strconv.Atoi(details["クリア回数"])
	songStatistics.PlayCount, err = strconv.Atoi(details["プレー回数"])
	songStatistics.BestScore, err = strconv.Atoi(details["ハイスコア"])
	songStatistics.Rank = details["ハイスコア時のダンスレベル"]
	songStatistics.Lamp = details["フルコンボ種別"]
	songStatistics.LastPlayed, err = time.ParseInLocation(timeFormat, details["最終プレー時間"], timeLocation)

	if err != nil {
		return
	}

	songStatistics.SongId = difficulty.SongId
	songStatistics.Mode = difficulty.Mode
	songStatistics.Difficulty = difficulty.Difficulty

	songStatistics.PlayerCode = playerCode
	return
}

func RecentScoresForClient(client util.EaClient, playerCode int) (scores []ddr_models.Score, err error) {
	document, err := recentScoresDocument(client)
	if err != nil {
		return
	}
	scores, err = recentScoresFromDocument(document, playerCode)
	return
}

// TODO: error handling
func recentScoresFromDocument(document *goquery.Document, playerCode int) (scores []ddr_models.Score, err error) {
	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
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
		difficulty, err := strconv.Atoi(href[len(href)-1:])
		if err != nil {
			glog.Errorf("strconv failed: %s\n", err.Error())
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
		t, err := time.ParseInLocation(timeFormat, timeSelection.Text(), timeLocation)
		if err != nil {
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

func WorkoutDataForClient(client util.EaClient, playerCode int) (workoutData []ddr_models.WorkoutData, err error) {
	document, err := workoutDocument(client)
	if err != nil {
		return
	}
	workoutData, err = workoutDataFromDocument(document, playerCode)
	return
}

func workoutDataFromDocument(document *goquery.Document, playerCode int) (workoutData []ddr_models.WorkoutData, err error) {
	format := "2006-01-02"
	loc, err := time.LoadLocation("Asia/Tokyo")

	table := document.Find("table#work_out_left")
	if table.Length() == 0 {
		err = fmt.Errorf("could not find work_out_left")
		return
	}

	tableBody := table.First().Find("tbody").First()
	if tableBody == nil {
		err = fmt.Errorf("could not find table body")
		return
	}

	tableBody.Find("tr").Each(func(i int, s *goquery.Selection) {
		if s.Find("td").Length() == 5 {
			wd := ddr_models.WorkoutData{}
			s.Find("td").Each(func(i int, dataSelection *goquery.Selection) {
				if i == 1 {
					t, err := time.ParseInLocation(format, dataSelection.Text(), loc)
					if err == nil {
						wd.Date = t
					}
				} else if i == 2 {
					numerical, err := regexp.Compile("[^0-9]+")
					if err != nil {
						panic(err)
					}
					numericStr := numerical.ReplaceAllString(dataSelection.Text(), "")
					wd.PlayCount, _ = strconv.Atoi(numericStr)
				} else if i == 3 {
					numerical, err := regexp.Compile("[^0-9.]+")
					if err != nil {
						glog.Errorf("regex failure! %s\n", err.Error())
						panic(err)
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
