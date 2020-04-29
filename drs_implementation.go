package main

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/eagate/drs"
	"github.com/chris-sg/eagate/util"
	"github.com/golang/glog"
)

func refreshDrsUser(client util.EaClient) (err error) {
	glog.Infof("Refreshing user %s\n", client.GetUsername())
	if !client.LoginState() {
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}

	dancerInfo, err := drs.LoadDancerInfo(client)
	if err != nil {
		glog.Errorf("%s\n", err.Error())
	}
	musicData, err := drs.LoadMusicData(client)
	if err != nil {
		glog.Errorf("%s\n", err.Error())
	}
	playHist, err := drs.LoadPlayHist(client)
	if err != nil {
		glog.Errorf("%s\n", err.Error())
	}

	d, err := json.MarshalIndent(dancerInfo, "", "  ")
	d, err = json.MarshalIndent(musicData, "", "  ")
	if err != nil {
		glog.Errorf("%s\n", err.Error())
	}
	glog.Info(string(d))
	d, err = json.MarshalIndent(playHist, "", "  ")

	playerDetails, profileSnapshot, songs, difficulties, playerSongStats, playerScores := drs.Transform(dancerInfo, musicData, playHist)

	d, err = json.MarshalIndent(playerDetails, "", "  ")
	d, err = json.MarshalIndent(profileSnapshot, "", "  ")
	d, err = json.MarshalIndent(songs, "", "  ")
	d, err = json.MarshalIndent(difficulties, "", "  ")
	if err != nil {
		glog.Errorf("%s\n", err.Error())
	}
	glog.Info(string(d))
	d, err = json.MarshalIndent(playerSongStats, "", "  ")
	d, err = json.MarshalIndent(playerScores, "", "  ")

	glog.Info(dancerInfo, musicData, playHist)
	return
}