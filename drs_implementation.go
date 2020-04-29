package main

import (
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
	musicData, err := drs.LoadMusicData(client)
	playHist, err := drs.LoadPlayHist(client)

	glog.Info(dancerInfo, musicData, playHist)
	return
}