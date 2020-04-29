package main

import (
	"fmt"
	"github.com/chris-sg/eagate/drs"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
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


	playerDetails, profileSnapshot, songs, difficulties, playerSongStats, playerScores := drs.Transform(dancerInfo, musicData, playHist)

	errs := eagate_db.GetDrsDb().AddPlayerDetails(playerDetails)
	PrintErrors("failed to add player details to db:", errs)
	errs = eagate_db.GetDrsDb().AddPlayerProfileSnapshot(profileSnapshot)
	PrintErrors("failed to add player profile snapshot to db:", errs)
	errs = eagate_db.GetDrsDb().AddSongs(songs)
	PrintErrors("failed to add songs to db:", errs)
	errs = eagate_db.GetDrsDb().AddDifficulties(difficulties)
	PrintErrors("failed to add difficulties to db:", errs)
	errs = eagate_db.GetDrsDb().AddPlayerSongStats(playerSongStats)
	PrintErrors("failed to add song stats to db:", errs)
	errs = eagate_db.GetDrsDb().AddPlayerScores(playerScores)
	PrintErrors("failed to add player scores to db:", errs)

	glog.Info(dancerInfo, musicData, playHist)
	return
}