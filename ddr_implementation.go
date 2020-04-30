package main

import (
	"fmt"
	"github.com/chris-sg/eagate/ddr"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_models/ddr_models"
	"github.com/chris-sg/eagate_models/user_models"
	"github.com/golang/glog"
)

// checkForNewSongs will load the song list from eagate and compare it
// against the song list from the database. Any songs only located in
// eagate will be returned.
func checkForNewSongs(client util.EaClient) (newSongs []string, errMsg string, err error) {
	glog.Infof("Checking for new songs as user %s\n", client.GetUsername())
	if !client.LoginState() {
		errMsg = "bad_cookie"
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	siteIds, err := ddr.SongIdsForClient(client)
	if err != nil {
		errMsg = "ddr_songid_fail"
		glog.Errorf("Failed to load eagate song ids: %s\n", err.Error())
		return
	}

	dbIds, errs := eagate_db.GetDdrDb().RetrieveSongIds()
	if PrintErrors("failed to load songs from db, errors:", errs) {
		errMsg = "ddr_songid_db_fail"
		err = fmt.Errorf("failed to load songs from db")
		return
	}

	glog.Infof("Comparing %d eagate songs against %d db songs for user %s\n", len(siteIds), len(dbIds), client.GetUsername())
	for i := len(siteIds)-1; i >= 0; i-- {
		for j, _ := range dbIds {
			if dbIds[j] == siteIds[i] {
				siteIds = append(siteIds[:i], siteIds[i+1:]...)
				dbIds = append(dbIds[:j], dbIds[j+1:]...)
				break
			}
		}
	}
	glog.Infof("%d new songs found on eagate\n", len(siteIds))
	newSongs = siteIds
	UpdateCookie(client)
	return
}

// updateNewSongs will load song data and song difficulties for the
// provided songIds slice. This intends to be used after checkForNewSongs
// to update the database.
func updateNewSongs(client util.EaClient, songIds []string) (errMsg string, err error) {
	glog.Infof("Updating %d new songs from user %s\n", len(songIds), client.GetUsername())
	if !client.LoginState() {
		errMsg = "bad_cookie"
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}

	songData, err := ddr.SongDataForClient(client, songIds)
	if err != nil {
		errMsg = "ddr_songdata_fail"
		glog.Errorf("Failed to get song data from client %s\n", client.GetUsername())
		return
	}
	glog.Infof("Update new songs got %d song data points\n", len(songData))
	errs := eagate_db.GetDdrDb().AddSongs(songData)
	if PrintErrors("failed to add songs to db:", errs) {
		errMsg = "ddr_addsong_fail"
		err = fmt.Errorf("failed to add songs to db")
		return
	}
	difficulties, err := ddr.SongDifficultiesForClient(client, songIds)
	if err != nil {
		errMsg = "ddr_songdiff_fail"
		glog.Errorf("Failed to get song difficulties from client %s\n", client.GetUsername())
		return
	}
	glog.Infof("Update new songs got %d song difficulty points\n", len(difficulties))
	errs = eagate_db.GetDdrDb().AddDifficulties(difficulties)
	if PrintErrors("failed to add difficulties to db:", errs) {
		errMsg = "ddr_adddiff_fail"
		err = fmt.Errorf("failed to add difficulties to db")
	}
	UpdateCookie(client)
	return
}

func refreshDdrUser(client util.EaClient) (errMsg string, err error) {
	glog.Infof("Refreshing user %s\n", client.GetUsername())
	if !client.LoginState() {
		errMsg = "bad_cookie"
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}

	pi, pc, err := ddr.PlayerInformationForClient(client)
	if err != nil {
		errMsg = "ddr_pi_fail"
		glog.Errorf("Failed to load player information for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}

	errs := eagate_db.GetDdrDb().AddPlayerDetails(pi)
	if PrintErrors("failed to add player details:", errs) {
		errMsg = "ddr_addpi_fail"
		err = fmt.Errorf("failed to add player details")
		return
	}
	errs = eagate_db.GetDdrDb().AddPlaycounts([]ddr_models.Playcount{pc})
	if PrintErrors("failed to add playcount:", errs) {
		errMsg = "ddr_addpc_fail"
		err = fmt.Errorf("failed to add playcount")
		return
	}

	newSongs, errMsg, err := checkForNewSongs(client)
	if err != nil {
		glog.Errorf("Failed to load new songs for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	if len(newSongs) > 0 {
		glog.Infof("Updating %d new songs for client %s\n", len(newSongs), client.GetUsername())
		errMsg, err = updateNewSongs(client, newSongs)
		if err != nil {
			glog.Errorf("Failed to update new songs for client %s: %s\n", client.GetUsername(), err.Error())
			return
		}
	}

	songIds, err := ddr.SongIdsForClient(client)
	if err != nil {
		errMsg = "ddr_songid_fail"
		glog.Errorf("Failed to load song ids for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	difficulties, err := ddr.SongDifficultiesForClient(client, songIds)
	if err != nil {
		errMsg = "ddr_songdiff_fail"
		glog.Errorf("Failed to load song difficulties for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	glog.Infof("Adding song difficulties to db client %s (%d difficulties)\n", client.GetUsername(), len(difficulties))
	errs = eagate_db.GetDdrDb().AddDifficulties(difficulties)
	if PrintErrors("failed to add difficulties to db:", errs) {
		errMsg = "ddr_adddiff_fail"
		err = fmt.Errorf("failed to add difficulties to db")
		return
	}

	difficulties, errs = eagate_db.GetDdrDb().RetrieveValidDifficulties()
	if PrintErrors("failed to retrieve difficulties from db:", errs) {
		errMsg = "ddr_retdiff_fail"
		err = fmt.Errorf("failed to retrieve difficulties from db")
		return
	}

	songStats, err := ddr.SongStatisticsForClient(client, difficulties, pi.Code)
	if err != nil {
		errMsg = "ddr_songstat_fail"
		glog.Errorf("Failed to load song statistics for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}
	glog.Infof("Adding song statistics to db client %s (%d statistics)", client.GetUsername(), len(songStats))
	eagate_db.GetDdrDb().AddSongStatistics(songStats)
	if PrintErrors("failed to add song statistics to db:", errs) {
		errMsg = "ddr_addsongstat_fail"
		err = fmt.Errorf("failed to add song statistics to db")
		return
	}

	recentScores, err := ddr.RecentScoresForClient(client, pi.Code)
	if err != nil {
		errMsg = "ddr_recent_fail"
		glog.Errorf("Failed to load recent scores for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}
	glog.Infof("Adding song scores to db client %s (%d scores)", client.GetUsername(), len(recentScores))
	errs = eagate_db.GetDdrDb().AddScores(recentScores)
	if PrintErrors("failed to add scores to db:", errs) {
		errMsg = "ddr_addscore_fail"
		err = fmt.Errorf("failed to add scores to db")
		return
	}

	workoutData, err := ddr.WorkoutDataForClient(client, pi.Code)
	if err != nil {
		errMsg = "ddr_wd_fail"
		glog.Errorf("Failed to load workout data for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}
	glog.Infof("Adding workout data to db client %s (%d datapoints)", client.GetUsername(), len(workoutData))
	errs = eagate_db.GetDdrDb().AddWorkoutData(workoutData)
	if PrintErrors("failed to add workout data to db:", errs) {
		errMsg = "ddr_addwd_fail"
		err = fmt.Errorf("failed to add workout data to db")
	}

	return
}

// updateSongStatistics will load the client's statistics for the given
// difficulties slice and update the statistics in the database.
func updateSongStatistics(client util.EaClient, difficulties []ddr_models.SongDifficulty) (errMsg string, err error) {
	glog.Infof("Updating song statistics for user %s (%d difficulties)\n", client.GetUsername(), len(difficulties))
	if !client.LoginState() {
		errMsg = "bad_cookie"
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	pi, _, err := ddr.PlayerInformationForClient(client)
	if err != nil {
		errMsg = "ddr_pi_fail"
		glog.Errorf("Failed to load player info for user %s: %s\n", client.GetUsername(), err.Error())
		return
	}

	stats, err := ddr.SongStatisticsForClient(client, difficulties, pi.Code)
	if err != nil {
		errMsg = "ddr_songstat_fail"
		glog.Errorf("Failed to load song statistics for user %s code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}

	errs := eagate_db.GetDdrDb().AddPlayerDetails(pi)
	if PrintErrors("failed to add player details to db:", errs) {
		errMsg = "ddr_addpi_fail"
		err = fmt.Errorf("failed to add player details to db")
		return
	}
	eagate_db.GetDdrDb().AddSongStatistics(stats)
	if PrintErrors("failed to add song statistics to db:", errs) {
		errMsg = "ddr_addsongstat_fail"
		err = fmt.Errorf("failed to add song statistics to db")
		return
	}
	UpdateCookie(client)
	return
}

// updatePlayerProfile will do a full update of the user's profile. This
// includes updating the player information, the playcount, adding the
// recent scores and updating song statistics.
// TODO: if the user has played more than 50 songs, this will not update
// unknown song statistics. This can currently still be achieved manually.
func updatePlayerProfile(user user_models.User, client util.EaClient) (errMsg string, err error) {
	glog.Infof("Updating player profile for %s\n", client.GetUsername())
	if !client.LoginState() {
		errMsg = "bad_cookie"
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	newPi, playcount, err := ddr.PlayerInformationForClient(client)
	if err != nil {
		glog.Errorf("Failed to load player info for user %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	newPi.EaGateUser = &user.Name
	dbPi, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByPlayerCode(newPi.Code)
	if PrintErrors("failed to retrieve player details:", errs) {
		errMsg = "ddr_retpi_fail"
		err = fmt.Errorf("failed to retrieve player details for code %d", newPi.Code)
		return
	}
	if dbPi.Code != 0 {
		glog.Infof("Player info found for code %d, will not refresh\n", newPi.Code)
		dbPlaycount, errs := eagate_db.GetDdrDb().RetrieveLatestPlaycountByPlayerCode(dbPi.Code)
		if PrintErrors("failed to retrieve latest playcount:", errs) {
			errMsg = "ddr_retpc_fail"
			err = fmt.Errorf("failed to retrieve latest playcount details for %d", dbPi.Code)
			return
		}
		if playcount.Playcount == dbPlaycount.Playcount {
			glog.Infof("Playcount for %d unchanged. Will not update", dbPi.Code)
			return
		}
	} else {
		glog.Infof("Player info not found for code %d, will refresh\n", newPi.Code)
		refreshDdrUser(client)
	}
	errs = eagate_db.GetDdrDb().AddPlayerDetails(newPi)
	if PrintErrors("failed to update player information:", errs) {
		errMsg = "ddr_addpi_fail"
		err = fmt.Errorf("failed to update player info for user %s", client.GetUsername())
		return
	}

	recentScores, err := ddr.RecentScoresForClient(client, newPi.Code)
	if err != nil {
		errMsg = "ddr_recent_fail"
		glog.Errorf("Failed to load recent scores for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
		return
	}
	if recentScores == nil {
		errMsg = "ddr_recent_fail"
		glog.Errorf("failed to load recent scores for code %d\n", newPi.Code)
		err = fmt.Errorf("failed to load recent scores for code %d", newPi.Code)
	}

	workoutData, err := ddr.WorkoutDataForClient(client, newPi.Code)
	if err != nil {
		errMsg = "ddr_wd_fail"
		glog.Errorf("Failed to load workout data for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
		return
	}

	recentSongIds := make([]string, 0)
	for _, score := range recentScores {
		found := false
		for _, addedId := range recentSongIds {
			if addedId == score.SongId {
				found = true
				break
			}
		}

		if !found {
			recentSongIds = append(recentSongIds, score.SongId)
		}
	}
	glog.Infof("Update found %d song ids to update\n", len(recentSongIds))

	dbSongIds, errs := eagate_db.GetDdrDb().RetrieveSongIds()
	if PrintErrors("failed to retrieve song ids from db:", errs) {
		errMsg = "ddr_songid_fail"
		err = fmt.Errorf("failed to retrieve song ids from db")
		return
	}
	for i := len(recentSongIds)-1; i >= 0; i-- {
		for j, _ := range dbSongIds {
			if recentSongIds[i] == dbSongIds[j] {
				recentSongIds = append(recentSongIds[:i], recentSongIds[i+1:]...)
				dbSongIds = append(dbSongIds[:j], dbSongIds[j+1:]...)
				break
			}
		}
	}

	if len(recentSongIds) > 0 {
		errMsg, err = updateNewSongs(client, recentSongIds)
		if err != nil {
			glog.Errorf("Failed to update new songs for client %s\n", client.GetUsername())
			return
		}
	}


	if recentScores != nil {
		errs = eagate_db.GetDdrDb().AddScores(recentScores)
		if PrintErrors("failed to add recent scores:", errs) {
			errMsg = "ddr_addscore_fail"
			err = fmt.Errorf("failed to add recent scores for user %s", client.GetUsername())
			return
		}

		songsToUpdate := make([]ddr_models.SongDifficulty, 0)

		for _, score := range recentScores {
			added := false
			for _, song := range songsToUpdate {
				if score.SongId == song.SongId && score.Mode == song.Mode && score.Difficulty == song.Difficulty {
					added = true
					break
				}
			}
			if !added {
				songsToUpdate = append(songsToUpdate, ddr_models.SongDifficulty{
					SongId:          score.SongId,
					Mode:            score.Mode,
					Difficulty:      score.Difficulty,
					DifficultyValue: 0,
				})
			}
		}

		statistics, err2 := ddr.SongStatisticsForClient(client, songsToUpdate, newPi.Code)
		err = err2
		if err != nil {
			errMsg = "ddr_songstat_fail"
			glog.Errorf("Failed to update song statistics for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
			return
		}
		errs = eagate_db.GetDdrDb().AddSongStatistics(statistics)
		if PrintErrors("failed to add song statistics to db:", errs) {
			errMsg = "ddr_addsongstat_fail"
			return
		}
		glog.Infof("Updated song statistics for user %s\n", client.GetUsername())
	}

	if workoutData != nil {
		errs = eagate_db.GetDdrDb().AddWorkoutData(workoutData)
		if PrintErrors("failed to add workout data to db:", errs) {
			errMsg = "ddr_addwd_fail"
			err = fmt.Errorf("failed to add workout data for user %s", client.GetUsername())
		}
	}

	errs = eagate_db.GetDdrDb().AddPlaycounts([]ddr_models.Playcount{playcount})
	if PrintErrors("failed to add playcounts:", errs) {
		errMsg = "ddr_addpc_fail"
		err = fmt.Errorf("failed to add playcount details for %s", client.GetUsername())
		return
	}

	glog.Infof("Profile update complete for user %s\n", client.GetUsername())
	UpdateCookie(client)

	return
}