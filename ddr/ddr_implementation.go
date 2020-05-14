package ddr

import (
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/ddr"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/chris-sg/bst_api/models/user_models"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
)

// checkForNewSongs will load the song list from eagate and compare it
// against the song list from the database. Any songs only located in
// eagate will be returned.
func checkForNewSongs(client util.EaClient) (newSongs []string, err bst_models.Error) {
	err = bst_models.ErrorOK
	glog.Infof("Checking for new songs as user %s\n", client.GetUsername())
	if !client.LoginState() {
		err = bst_models.ErrorBadCookie
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		return
	}
	siteIds, err := ddr.SongIdsForClient(client)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load eagate song ids: %s\n", err.Message)
		return
	}

	dbIds, errs := db.GetDdrDb().RetrieveSongIds()
	if utilities.PrintErrors("failed to load songs from db, errors:", errs) {
		err = bst_models.ErrorDdrSongIdsDbRead
		return
	}

	glog.Infof("Comparing %d eagate songs against %d db songs for user %s\n", len(siteIds), len(dbIds), client.GetUsername())
	for i := len(siteIds)-1; i >= 0; i-- {
		for j := range dbIds {
			if dbIds[j] == siteIds[i] {
				siteIds = append(siteIds[:i], siteIds[i+1:]...)
				dbIds = append(dbIds[:j], dbIds[j+1:]...)
				break
			}
		}
	}
	glog.Infof("%d new songs found on eagate\n", len(siteIds))
	newSongs = siteIds
	utilities.UpdateCookie(client)
	return
}

// updateNewSongs will load song data and song difficulties for the
// provided songIds slice. This intends to be used after checkForNewSongs
// to update the database.
func updateNewSongs(client util.EaClient, songIds []string) bst_models.Error {
	glog.Infof("Updating %d new songs from user %s\n", len(songIds), client.GetUsername())
	if !client.LoginState() {
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		return bst_models.ErrorBadCookie
	}

	songData, err := ddr.SongDataForClient(client, songIds)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to get song data from client %s\n", client.GetUsername())
		return err
	}
	glog.Infof("Update new songs got %d song data points\n", len(songData))
	errs := db.GetDdrDb().AddSongs(songData)
	if utilities.PrintErrors("failed to add songs to db:", errs) {
		return bst_models.ErrorDdrSongDataDbWrite
	}
	difficulties, err := ddr.SongDifficultiesForClient(client, songIds)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to get song difficulties from client %s\n", client.GetUsername())
		return err
	}
	glog.Infof("Update new songs got %d song difficulty points\n", len(difficulties))
	errs = db.GetDdrDb().AddDifficulties(difficulties)
	if utilities.PrintErrors("failed to add difficulties to db:", errs) {
		return bst_models.ErrorDdrSongDifficultiesDbWrite
	}
	utilities.UpdateCookie(client)
	return bst_models.ErrorOK
}

func refreshDdrUser(client util.EaClient) (err bst_models.Error) {
	err = bst_models.ErrorOK
	glog.Infof("Refreshing user %s\n", client.GetUsername())
	if !client.LoginState() {
		err = bst_models.ErrorBadCookie
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		return
	}

	pi, pc, err := ddr.PlayerInformationForClient(client)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load player information for client %s: %s\n", client.GetUsername(), err.Message)
		return
	}

	errs := db.GetDdrDb().AddPlayerDetails(pi)
	if utilities.PrintErrors("failed to add player details:", errs) {
		err = bst_models.ErrorDdrPlayerInfoDbWrite
		return
	}
	errs = db.GetDdrDb().AddPlaycounts([]ddr_models.Playcount{pc})
	if utilities.PrintErrors("failed to add playcount:", errs) {
		err = bst_models.ErrorDdrPlayerInfoDbWrite
		return
	}

	newSongs, err := checkForNewSongs(client)
	if !err.Equals(bst_models.ErrorOK) {
		return
	}
	if len(newSongs) > 0 {
		glog.Infof("Updating %d new songs for client %s\n", len(newSongs), client.GetUsername())
		err = updateNewSongs(client, newSongs)
		if !err.Equals(bst_models.ErrorOK) {
			glog.Errorf("Failed to update new songs for client %s: %s\n", client.GetUsername(), err.Message)
			return
		}
	}

	songIds, err := ddr.SongIdsForClient(client)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load song ids for client %s: %s\n", client.GetUsername(), err.Message)
		return
	}
	difficulties, err := ddr.SongDifficultiesForClient(client, songIds)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load song difficulties for client %s: %s\n", client.GetUsername(), err.Message)
		return
	}
	glog.Infof("Adding song difficulties to db client %s (%d difficulties)\n", client.GetUsername(), len(difficulties))
	errs = db.GetDdrDb().AddDifficulties(difficulties)
	if utilities.PrintErrors("failed to add difficulties to db:", errs) {
		err = bst_models.ErrorDdrSongDifficultiesDbWrite
		return
	}

	difficulties, errs = db.GetDdrDb().RetrieveValidDifficulties()
	if utilities.PrintErrors("failed to retrieve difficulties from db:", errs) {
		err = bst_models.ErrorDdrSongDifficultiesDbRead
		return
	}

	songStats, err := ddr.SongStatisticsForClient(client, difficulties, pi.Code)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load song statistics for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Message)
		return
	}
	glog.Infof("Adding song statistics to db client %s (%d statistics)", client.GetUsername(), len(songStats))
	db.GetDdrDb().AddSongStatistics(songStats)
	if utilities.PrintErrors("failed to add song statistics to db:", errs) {
		err = bst_models.ErrorDdrStatsDbWrite
		return
	}

	recentScores, err := ddr.RecentScoresForClient(client, pi.Code)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load recent scores for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Message)
		return
	}
	glog.Infof("Adding song scores to db client %s (%d scores)", client.GetUsername(), len(recentScores))
	errs = db.GetDdrDb().AddScores(recentScores)
	if utilities.PrintErrors("failed to add scores to db:", errs) {
		err = bst_models.ErrorDdrStatsDbWrite
		return
	}

	workoutData, err := ddr.WorkoutDataForClient(client, pi.Code)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load workout data for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Message)
		return
	}
	glog.Infof("Adding workout data to db client %s (%d datapoints)", client.GetUsername(), len(workoutData))
	errs = db.GetDdrDb().AddWorkoutData(workoutData)
	if utilities.PrintErrors("failed to add workout data to db:", errs) {
		err = bst_models.ErrorDdrStatsDbWrite
	}

	return
}

// updatePlayerProfile will do a full update of the user's profile. This
// includes updating the player information, the playcount, adding the
// recent scores and updating song statistics.
// TODO: if the user has played more than 50 songs, this will not update
// unknown song statistics. This can currently still be achieved manually.
func updatePlayerProfile(user user_models.User, client util.EaClient) (err bst_models.Error) {
	err = bst_models.ErrorOK
	glog.Infof("Updating player profile for %s\n", client.GetUsername())
	if !client.LoginState() {
		err = bst_models.ErrorBadCookie
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		return
	}
	newPi, playcount, err := ddr.PlayerInformationForClient(client)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load player info for user %s: %s\n", client.GetUsername(), err.Message)
		return
	}
	newPi.EaGateUser = &user.Name
	dbPi, errs := db.GetDdrDb().RetrievePlayerDetailsByPlayerCode(newPi.Code)
	if utilities.PrintErrors("failed to retrieve player details:", errs) {
		err = bst_models.ErrorDdrStatsDbRead
		return
	}
	if dbPi.Code != 0 {
		glog.Infof("Player info found for code %d, will not refresh\n", newPi.Code)
		dbPlaycount, errs := db.GetDdrDb().RetrieveLatestPlaycountByPlayerCode(dbPi.Code)
		if utilities.PrintErrors("failed to retrieve latest playcount:", errs) {
			err = bst_models.ErrorDdrStatsDbRead
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
	errs = db.GetDdrDb().AddPlayerDetails(newPi)
	if utilities.PrintErrors("failed to update player information:", errs) {
		err = bst_models.ErrorDdrStatsDbWrite
		return
	}

	recentScores, err := ddr.RecentScoresForClient(client, newPi.Code)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load recent scores for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Message)
		return
	}
	if recentScores == nil {
		err = bst_models.ErrorDdrStats
		glog.Errorf("failed to load recent scores for code %d\n", newPi.Code)
		return
	}

	workoutData, err := ddr.WorkoutDataForClient(client, newPi.Code)
	if !err.Equals(bst_models.ErrorOK) {
		glog.Errorf("Failed to load workout data for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Message)
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

	dbSongIds, errs := db.GetDdrDb().RetrieveSongIds()
	if utilities.PrintErrors("failed to retrieve song ids from db:", errs) {
		err = bst_models.ErrorDdrStatsDbRead
		return
	}
	for i := len(recentSongIds)-1; i >= 0; i-- {
		for j := range dbSongIds {
			if recentSongIds[i] == dbSongIds[j] {
				recentSongIds = append(recentSongIds[:i], recentSongIds[i+1:]...)
				dbSongIds = append(dbSongIds[:j], dbSongIds[j+1:]...)
				break
			}
		}
	}

	if len(recentSongIds) > 0 {
		err = updateNewSongs(client, recentSongIds)
		if !err.Equals(bst_models.ErrorOK) {
			glog.Errorf("Failed to update new songs for client %s\n", client.GetUsername())
			return
		}
	}


	if recentScores != nil {
		errs = db.GetDdrDb().AddScores(recentScores)
		if utilities.PrintErrors("failed to add recent scores:", errs) {
			err = bst_models.ErrorDdrStatsDbWrite
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
		if !err.Equals(bst_models.ErrorOK) {
			glog.Errorf("Failed to update song statistics for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Message)
			return
		}
		errs = db.GetDdrDb().AddSongStatistics(statistics)
		if utilities.PrintErrors("failed to add song statistics to db:", errs) {
			err = bst_models.ErrorDdrStatsDbWrite
			return
		}
		glog.Infof("Updated song statistics for user %s\n", client.GetUsername())
	}

	if workoutData != nil {
		errs = db.GetDdrDb().AddWorkoutData(workoutData)
		if utilities.PrintErrors("failed to add workout data to db:", errs) {
			err = bst_models.ErrorDdrStatsDbWrite
			return
		}
	}

	errs = db.GetDdrDb().AddPlaycounts([]ddr_models.Playcount{playcount})
	if utilities.PrintErrors("failed to add playcounts:", errs) {
		err = bst_models.ErrorDdrStatsDbWrite
		return
	}

	glog.Infof("Profile update complete for user %s\n", client.GetUsername())
	utilities.UpdateCookie(client)

	return
}