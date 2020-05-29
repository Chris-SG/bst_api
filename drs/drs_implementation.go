package drs

import (
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/drs"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/drs_models"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
)

func refreshDrsUser(client util.EaClient) (err bst_models.Error) {
	err = bst_models.ErrorOK
	glog.Infof("Refreshing user %s\n", client.GetUserModel().Name)
	if !client.LoginState() {
		err = bst_models.ErrorBadCookie
		return
	}

	dancerInfo, err := drs.LoadDancerInfo(client)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	musicData, err := drs.LoadMusicData(client)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	playHist, err := drs.LoadPlayHist(client)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}


	playerDetails, profileSnapshot, songs, difficulties, playerSongStats, playerScores := drs.Transform(dancerInfo, musicData, playHist)
	user := client.GetUserModel().Name
	if len(user) > 0 {
		playerDetails.EaGateUser = &user
	}

	errs := db.GetDrsDb().AddPlayerDetails(playerDetails)
	utilities.PrintErrors("failed to add player details to db:", errs)
	errs = db.GetDrsDb().AddPlayerProfileSnapshot(profileSnapshot)
	utilities.PrintErrors("failed to add player profile snapshot to db:", errs)
	errs = db.GetDrsDb().AddSongs(songs)
	utilities.PrintErrors("failed to add songs to db:", errs)
	errs = db.GetDrsDb().AddDifficulties(difficulties)
	utilities.PrintErrors("failed to add difficulties to db:", errs)
	errs = db.GetDrsDb().AddPlayerSongStats(playerSongStats)
	utilities.PrintErrors("failed to add song stats to db:", errs)
	errs = db.GetDrsDb().AddPlayerScores(playerScores)
	utilities.PrintErrors("failed to add player scores to db:", errs)

	return
}

func retrieveDrsPlayerDetails(eaUser string) (details drs_models.PlayerDetails, err bst_models.Error) {
	err = bst_models.ErrorOK
	details, errs := db.GetDrsDb().RetrievePlayerDetailsByEaGateUser(eaUser)
	if utilities.PrintErrors("failed to retrieve user:", errs) {
		err = bst_models.ErrorDrsPlayerInfoDbRead
	}
	return
}

func retrieveDrsSongStats(code int) (songStats []drs_models.PlayerSongStats, err bst_models.Error) {
	songStats, errs := db.GetDrsDb().RetrieveSongStatisticsByPlayerCode(code)
	if utilities.PrintErrors("failed to retrieve user:", errs) {
		err = bst_models.ErrorDrsSongDataDbRead
	}
	return
}