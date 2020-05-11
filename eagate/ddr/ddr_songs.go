package ddr

import (
	"encoding/base64"
	"fmt"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/chris-sg/bst_api/eagate/util"
)

func SongIdsForClient(client util.EaClient) (songIds []string, err error) {
	mtx := &sync.Mutex{}

	musicDataDoc, err := musicDataSingleDocument(client, 0)
	if err != nil {
		return
	}
	pageCount := pageCountFromMusicDataDocument(musicDataDoc)

	errCount := 0

	wg := new(sync.WaitGroup)
	wg.Add(pageCount)

	for idx := 0; idx < pageCount; idx++ {
		go func(page int) {
			defer wg.Done()

			musicDataDoc, err := musicDataSingleDocument(client, page)
			if err != nil {
				errCount++
				glog.Errorf("failed to load musicDataSingleDocument for user %s page %d: %s\n", client.GetUsername(), page, err.Error())
				return
			}

			pageSongIds := songIdsFromMusicDataDocument(musicDataDoc)

			mtx.Lock()
			defer mtx.Unlock()
			songIds = append(songIds, pageSongIds...)
		}(idx)
	}

	wg.Wait()
	glog.Infof("loaded %d song ids on user %s\n", len(songIds), client.GetUsername())

	if errCount != 0 {
		err = fmt.Errorf("failed to load %d/%d pages", errCount, pageCount)
	}

	return
}

func songIdsFromMusicDataDocument(document *goquery.Document) (songIds []string) {
	const songDetailBaseUri = "/game/ddr/ddra20/p/playdata/music_detail.html?index="
	document.Find("tr.data").Each(func(i int, s *goquery.Selection) {
		aElement := s.Find("a").First()
		href, exists := aElement.Attr("href")
		if exists {
			id := strings.Replace(href, songDetailBaseUri, "", -1)
			songIds = append(songIds, id)
		}
	})
	return
}

func pageCountFromMusicDataDocument(document *goquery.Document) (pageCount int) {
	pageCount = document.Find("div#paging_box").First().Find("div.page_num").Length()
	return
}

func SongDataForClient(client util.EaClient, songIds []string) (songs []ddr_models.Song, err error) {
	mtx := &sync.Mutex{}

	wg := new(sync.WaitGroup)
	wg.Add(len(songIds))

	errCount := 0

	for _, id := range songIds {
		go func(songId string) {
			defer wg.Done()
			document, err := musicDetailDocument(client, songId)
			if err != nil {
				glog.Errorf("failed to get document for song id %s: %s", songId, err.Error())
				errCount++
				return
			}
			song := songDataFromDocument(document, songId)

			mtx.Lock()
			defer mtx.Unlock()
			songs = append(songs, song)
		}(id)
	}

	wg.Wait()
	glog.Infof("loaded %d song data on user %s\n", len(songs), client.GetUsername())

	if errCount > 0 {
		glog.Warningf("failed %d/%d song ids for song data (user %s)\n", errCount, len(songIds), client.GetUsername())
		err = fmt.Errorf("failed to load data for %d/%d songs", errCount, len(songIds))
	}

	return
}

func songDataFromDocument(document *goquery.Document, songId string) (song ddr_models.Song) {
	song.Id = songId
	document.Find("table#music_info").First().Find("td").Each(func(i int, s *goquery.Selection) {
		img := s.Find("img")
		if img.Length() == 0 {
			html, _ := s.Html()
			songDataPair := strings.Split(html, "<br/>")
			song.Name = songDataPair[0]
			song.Artist = songDataPair[1]
		} else {
			imgPath, exists := img.First().Attr("src")
			if exists {
				imgUrl := util.BuildEaURI(imgPath)
				imgData, err := http.Get(imgUrl)
				defer imgData.Body.Close()
				if err == nil {
					body, err := ioutil.ReadAll(imgData.Body)
					if err == nil {
						song.Image = base64.StdEncoding.EncodeToString(body)
					}
				}
			}
		}
	})
	return
}

func SongDifficultiesForClient(client util.EaClient, songIds []string) (difficulties []ddr_models.SongDifficulty, err error) {
	mtx := &sync.Mutex{}

	wg := new(sync.WaitGroup)
	wg.Add(len(songIds))

	errCount := 0

	for _, id := range songIds {
		go func(songId string) {
			defer wg.Done()
			document, err := musicDetailDocument(client, songId)
			if err != nil {
				glog.Errorf("failed to get document for song id %s: %s", songId, err.Error())
				errCount++
				return
			}
			songDifficulties := songDifficultiesFromDocument(document, songId)

			mtx.Lock()
			defer mtx.Unlock()
			difficulties = append(difficulties, songDifficulties...)
		}(id)
	}

	wg.Wait()
	glog.Infof("loaded %d song difficulties for %d songs on user %s\n", len(difficulties), len(songIds), client.GetUsername())

	if errCount > 0 {
		glog.Warningf("failed %d/%d song ids for song data (user %s)\n", errCount, len(songIds), client.GetUsername())
		err = fmt.Errorf("failed to load data for %d/%d  songs", errCount, len(songIds))
	}

	return
}

func songDifficultiesFromDocument(document *goquery.Document, songId string) (songDifficulties []ddr_models.SongDifficulty) {
	single := document.Find("div#single")
	double := document.Find("div#double")

	single.Find("li.step").Each(func(i int, s *goquery.Selection) {
		img, exists := s.Find("img").Attr("src")
		if exists {
			imgExp := regexp.MustCompile(`songdetails_level_[0-9]*\.png`)
			lvlExp := regexp.MustCompile("[^0-9]+")
			s := imgExp.FindString(img)
			s = lvlExp.ReplaceAllString(s, "")
			v, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				v = -1
			}

			songDifficulties = append(songDifficulties, ddr_models.SongDifficulty{
				SongId:          songId,
				Mode: ddr_models.Single.String(),
				Difficulty:    ddr_models.Difficulty(i).String(),
				DifficultyValue: int16(v),
			})
		}
	})

	double.Find("li.step").Each(func(i int, s *goquery.Selection) {
		img, exists := s.Find("img").Attr("src")
		if exists {
			imgExp := regexp.MustCompile(`songdetails_level_[0-9]*\.png`)
			lvlExp := regexp.MustCompile("[^0-9]+")
			s := imgExp.FindString(img)
			s = lvlExp.ReplaceAllString(s, "")
			v, err := strconv.ParseInt(s, 10, 8)
			if err != nil {
				v = -1
			}

			songDifficulties = append(songDifficulties, ddr_models.SongDifficulty{
				SongId:          songId,
				Mode: ddr_models.Double.String(),
				Difficulty:    ddr_models.Difficulty(i).String(),
				DifficultyValue: int16(v),
			})
		}
	})
	return
}