package drs

import (
	"fmt"
	"github.com/chris-sg/bst_api/utilities"
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
	utilities.PrintErrors("failed to add player details to db:", errs)
	errs = eagate_db.GetDrsDb().AddPlayerProfileSnapshot(profileSnapshot)
	utilities.PrintErrors("failed to add player profile snapshot to db:", errs)
	errs = eagate_db.GetDrsDb().AddSongs(songs)
	utilities.PrintErrors("failed to add songs to db:", errs)
	errs = eagate_db.GetDrsDb().AddDifficulties(difficulties)
	utilities.PrintErrors("failed to add difficulties to db:", errs)
	errs = eagate_db.GetDrsDb().AddPlayerSongStats(playerSongStats)
	utilities.PrintErrors("failed to add song stats to db:", errs)
	errs = eagate_db.GetDrsDb().AddPlayerScores(playerScores)
	utilities.PrintErrors("failed to add player scores to db:", errs)

	return
}