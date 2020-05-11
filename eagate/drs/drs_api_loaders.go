package drs

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/drs_models"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func LoadDancerInfo(client util.EaClient) (dancerInfo drs_models.DancerInfo, err error) {
	const dancerInfoSingleResource = "/game/dan/1st/json/pdata_getdata.html"
	dancerInfoURI := util.BuildEaURI(dancerInfoSingleResource)

	form := url.Values{}
	form.Add("service_kind", "dancer_info")
	form.Add("pdata_kind", "dancer_info")

	req, err := http.NewRequest(http.MethodPost, dancerInfoURI, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	glog.Infof("retrieving resource %s\n", dancerInfoURI)
	res, err := client.Client.Do(req)

	if err != nil {
		glog.Errorf("failed to get resource %s: %s\n", dancerInfoURI, err.Error())
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	contentType, ok := res.Header["Content-Type"]
	if ok && len(contentType) > 0 {
		if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
			body = util.ShiftJISBytesToUTF8Bytes(body)
		}
	}

	err = json.Unmarshal(body, &dancerInfo)
	return
}

func LoadMusicData(client util.EaClient) (musicData drs_models.MusicData, err error) {
	const musicDataSingleResource = "/game/dan/1st/json/pdata_getdata.html"
	musicDataURI := util.BuildEaURI(musicDataSingleResource)

	form := url.Values{}
	form.Add("service_kind", "music_data")
	form.Add("pdata_kind", "music_data")

	req, err := http.NewRequest(http.MethodPost, musicDataURI, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	glog.Infof("retrieving resource %s\n", musicDataURI)
	res, err := client.Client.Do(req)

	if err != nil {
		glog.Errorf("failed to get resource %s: %s\n", musicDataURI, err.Error())
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	contentType, ok := res.Header["Content-Type"]
	if ok && len(contentType) > 0 {
		if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
			body = util.ShiftJISBytesToUTF8Bytes(body)
		}
	}

	err = json.Unmarshal(body, &musicData)
	return
}

func LoadPlayHist(client util.EaClient) (playHist drs_models.PlayHist, err error) {
	const playHistSingleResource = "/game/dan/1st/json/pdata_getdata.html"
	playHistURI := util.BuildEaURI(playHistSingleResource)

	form := url.Values{}
	form.Add("service_kind", "play_hist")
	form.Add("pdata_kind", "play_hist")

	req, err := http.NewRequest(http.MethodPost, playHistURI, strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	glog.Infof("retrieving resource %s\n", playHistURI)
	res, err := client.Client.Do(req)

	if err != nil {
		glog.Errorf("failed to get resource %s: %s\n", playHistURI, err.Error())
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	contentType, ok := res.Header["Content-Type"]
	if ok && len(contentType) > 0 {
		if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
			body = util.ShiftJISBytesToUTF8Bytes(body)
		}
	}

	err = json.Unmarshal(body, &playHist)
	return
}
