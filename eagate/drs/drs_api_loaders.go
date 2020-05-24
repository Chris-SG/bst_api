package drs

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/drs_models"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func LoadDancerInfo(client util.EaClient) (dancerInfo drs_models.DancerInfo, err bst_models.Error) {
	err = bst_models.ErrorOK
	const dancerInfoSingleResource = "/game/dan/1st/json/pdata_getdata.html"
	dancerInfoURI := util.BuildEaURI(dancerInfoSingleResource)

	form := url.Values{}
	form.Add("service_kind", "dancer_info")
	form.Add("pdata_kind", "dancer_info")

	req, e := http.NewRequest(http.MethodPost, dancerInfoURI, strings.NewReader(form.Encode()))
	if e != nil {
		glog.Errorf("failed to get resource %s: %s\n", dancerInfoURI, e.Error())
		err = bst_models.ErrorCreateRequest
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	glog.Infof("retrieving resource %s\n", dancerInfoURI)
	res, e := client.Client.Do(req)

	if e != nil {
		glog.Errorf("failed to get resource %s: %s\n", dancerInfoURI, e.Error())
		err = bst_models.ErrorClientRequest
		return
	}
	defer res.Body.Close()
	body, e := ioutil.ReadAll(res.Body)
	if e != nil {
		glog.Errorf("failed to read response %s: %s\n", dancerInfoURI, e.Error())
		err = bst_models.ErrorClientResponse
		return
	}

	contentType, ok := res.Header["Content-Type"]
	if ok && len(contentType) > 0 {
		if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
			body = util.ShiftJISBytesToUTF8Bytes(body)
		}
	}

	e = json.Unmarshal(body, &dancerInfo)
	if e != nil {
		glog.Errorf("failed to decode json: %s", e.Error())
		err = bst_models.ErrorJsonDecode
	}
	return
}

func LoadMusicData(client util.EaClient) (musicData drs_models.MusicData, err bst_models.Error) {
	err = bst_models.ErrorOK
	const musicDataSingleResource = "/game/dan/1st/json/pdata_getdata.html"
	musicDataURI := util.BuildEaURI(musicDataSingleResource)

	form := url.Values{}
	form.Add("service_kind", "music_data")
	form.Add("pdata_kind", "music_data")

	req, e := http.NewRequest(http.MethodPost, musicDataURI, strings.NewReader(form.Encode()))
	if e != nil {
		glog.Errorf("failed to get resource %s: %s\n", musicDataURI, e.Error())
		err = bst_models.ErrorCreateRequest
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	glog.Infof("retrieving resource %s\n", musicDataURI)
	res, e := client.Client.Do(req)

	if e != nil {
		glog.Errorf("failed to get resource %s: %s\n", musicDataURI, e.Error())
		err = bst_models.ErrorClientRequest
		return
	}
	defer res.Body.Close()
	body, e := ioutil.ReadAll(res.Body)
	if e != nil {
		glog.Errorf("failed to read response %s: %s\n", musicDataURI, e.Error())
		err = bst_models.ErrorClientResponse
		return
	}

	contentType, ok := res.Header["Content-Type"]
	if ok && len(contentType) > 0 {
		if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
			body = util.ShiftJISBytesToUTF8Bytes(body)
		}
	}

	e = json.Unmarshal(body, &musicData)
	if e != nil {
		glog.Errorf("failed to decode json: %s", e.Error())
		err = bst_models.ErrorJsonDecode
	}
	return
}

func LoadPlayHist(client util.EaClient) (playHist drs_models.PlayHist, err bst_models.Error) {
	err = bst_models.ErrorOK
	const playHistSingleResource = "/game/dan/1st/json/pdata_getdata.html"
	playHistURI := util.BuildEaURI(playHistSingleResource)

	form := url.Values{}
	form.Add("service_kind", "play_hist")
	form.Add("pdata_kind", "play_hist")

	req, e := http.NewRequest(http.MethodPost, playHistURI, strings.NewReader(form.Encode()))
	if e != nil {
		glog.Errorf("failed to get resource %s: %s\n", playHistURI, e.Error())
		err = bst_models.ErrorCreateRequest
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	glog.Infof("retrieving resource %s\n", playHistURI)
	res, e := client.Client.Do(req)

	if e != nil {
		glog.Errorf("failed to get resource %s: %s\n", playHistURI, e.Error())
		err = bst_models.ErrorClientRequest
		return
	}
	defer res.Body.Close()
	body, e := ioutil.ReadAll(res.Body)
	if e != nil {
		glog.Errorf("failed to read response %s: %s\n", playHistURI, e.Error())
		err = bst_models.ErrorClientResponse
		return
	}

	contentType, ok := res.Header["Content-Type"]
	if ok && len(contentType) > 0 {
		if strings.Contains(res.Header["Content-Type"][0], "Windows-31J") {
			body = util.ShiftJISBytesToUTF8Bytes(body)
		}
	}

	e = json.Unmarshal(body, &playHist)
	if e != nil {
		glog.Errorf("failed to decode json: %s", e.Error())
		err = bst_models.ErrorJsonDecode
	}
	return
}
