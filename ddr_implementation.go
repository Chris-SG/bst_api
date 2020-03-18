package main

import (
	"fmt"
	"github.com/chris-sg/eagate/ddr"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/ddr_db"
	"github.com/chris-sg/eagate_models/ddr_models"
	"github.com/chris-sg/eagate_models/user_models"
	"github.com/golang/glog"
)

// checkForNewSongs will load the song list from eagate and compare it
// against the song list from the database. Any songs only located in
// eagate will be returned.
func checkForNewSongs(client util.EaClient) (newSongs []string, err error) {
	glog.Infof("Checking for new songs as user %s\n", client.GetUsername())
	if !client.LoginState() {
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	siteIds, err := ddr.SongIds(client)
	if err != nil {
		glog.Errorf("Failed to load eagate song ids: %s\n", err.Error())
		return
	}
	db, err := eagate_db.GetDb()
	if err != nil {
		glog.Errorf("Failed to load db: %s\n", err.Error())
		return
	}
	dbIds := ddr_db.RetrieveSongIds(db)

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
func updateNewSongs(client util.EaClient, songIds []string) error {
	glog.Infof("Updating %d new songs from user %s\n", len(songIds), client.GetUsername())
	if !client.LoginState() {
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err := fmt.Errorf("user not logged into eagate")
		return err
	}
	db, _ := eagate_db.GetDb()
	songData, err := ddr.SongData(client, songIds)
	if err != nil {
		glog.Errorf("Failed to get song data from client %s\n", client.GetUsername())
		return err
	}
	glog.Infof("Update new songs got %d song data points\n", len(songData))
	err = ddr_db.AddSongs(db, songData)
	if err != nil {
		glog.Errorf("Update failed to add songs to db: %s\n", err.Error())
		return err
	}
	difficulties, err := ddr.SongDifficulties(client, songIds)
	if err != nil {
		glog.Errorf("Failed to get song difficulties from client %s\n", client.GetUsername())
		return err
	}
	glog.Infof("Update new songs got %d song difficulty points\n", len(difficulties))
	err = ddr_db.AddSongDifficulties(db, difficulties)
	if err != nil {
		glog.Errorf("Update failed to add difficulties to db: %s\n", err.Error())
		return err
	}
	UpdateCookie(client)
	return nil
}

func refreshDdrUser(client util.EaClient) (err error) {
	glog.Infof("Refreshing user %s\n", client.GetUsername())
	if !client.LoginState() {
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		glog.Errorf("Failed to load db: %s\n", err.Error())
		return
	}

	pi, pc, err := ddr.PlayerInformation(client)
	if err != nil {
		glog.Errorf("Failed to load player information for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	ddr_db.AddPlayerDetails(db, *pi)
	ddr_db.AddPlaycountDetails(db, *pc)

	newSongs, err := checkForNewSongs(client)
	if err != nil {
		glog.Errorf("Failed to load new songs for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	if len(newSongs) > 0 {
		glog.Infof("Updating %d new songs for client %s\n", len(newSongs), client.GetUsername())
		err = updateNewSongs(client, newSongs)
		if err != nil {
			glog.Errorf("Failed to update new songs for client %s: %s\n", client.GetUsername(), err.Error())
			return
		}
	}

	songIds, err := ddr.SongIds(client)
	if err != nil {
		glog.Errorf("Failed to load song ids for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	difficulties, err := ddr.SongDifficulties(client, songIds)
	if err != nil {
		glog.Errorf("Failed to load song difficulties for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	glog.Infof("Adding song difficulties to db client %s (%d difficulties)\n", client.GetUsername(), len(difficulties))
	err = ddr_db.AddSongDifficulties(db, difficulties)
	if err != nil {
		glog.Errorf("Failed to add song difficulties to db for client %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	difficulties = ddr_db.RetrieveValidSongDifficulties(db)

	songStats, err := ddr.SongStatistics(client, difficulties, pi.Code)
	if err != nil {
		glog.Errorf("Failed to load song statistics for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}
	glog.Infof("Adding song statistics to db client %s (%d statistics)", client.GetUsername(), len(songStats))
	err = ddr_db.AddSongStatistics(db, songStats, pi.Code)
	if err != nil {
		glog.Errorf("Failed to add song statistics to db for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}

	recentScores, err := ddr.RecentScores(client, pi.Code)
	if err != nil {
		glog.Errorf("Failed to load recent scores for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}
	glog.Infof("Adding song scores to db client %s (%d scores)", client.GetUsername(), len(*recentScores))
	err = ddr_db.AddScores(db, *recentScores)
	if err != nil {
		glog.Errorf("Failed to add recent scores for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}

	workoutData, err := ddr.WorkoutData(client, pi.Code)
	if err != nil {
		glog.Errorf("Failed to load workout data for client %s, code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}
	glog.Infof("Adding workout data to db client %s (%d datapoints)", client.GetUsername(), len(workoutData))
	ddr_db.AddWorkoutData(db, workoutData)

	return
}

// updateSongStatistics will load the client's statistics for the given
// difficulties slice and update the statistics in the database.
func updateSongStatistics(client util.EaClient, difficulties []ddr_models.SongDifficulty) (err error) {
	glog.Infof("Updating song statistics for user %s (%d difficulties)\n", client.GetUsername(), len(difficulties))
	if !client.LoginState() {
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	pi, _, err := ddr.PlayerInformation(client)
	if err != nil {
		glog.Errorf("Failed to load player info for user %s: %s\n", client.GetUsername(), err.Error())
		return
	}

	stats, err := ddr.SongStatistics(client, difficulties, pi.Code)
	if err != nil {
		glog.Errorf("Failed to load song statistics for user %s code %d: %s\n", client.GetUsername(), pi.Code, err.Error())
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		glog.Errorf("Failed to load db: %s\n", err.Error())
		return
	}
	err = ddr_db.AddPlayerDetails(db, *pi)
	if err != nil {
		glog.Errorf("Failed to add player details for user %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	ddr_db.AddSongStatistics(db, stats, pi.Code)
	if err != nil {
		glog.Errorf("Failed to add %d song statistics for user %s: %s\n", len(stats), client.GetUsername(), err.Error())
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
func updatePlayerProfile(user user_models.User, client util.EaClient) (err error) {
	glog.Infof("Updating player profile for %s\n", client.GetUsername())
	if !client.LoginState() {
		glog.Errorf("Client %s not logged into eagate\n", client.GetUsername())
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	db, err := eagate_db.GetDb()
	if err != nil {
		glog.Errorf("Failed to load db: %s\n", err.Error())
		return
	}
	newPi, playcount, err := ddr.PlayerInformation(client)
	if err != nil {
		glog.Errorf("Failed to load player info for user %s: %s\n", client.GetUsername(), err.Error())
		return
	}
	newPi.EaGateUser = &user.Name
	dbPi, _ := ddr_db.RetrieveDdrPlayerDetailsByCode(db, newPi.Code)
	if dbPi != nil {
		glog.Infof("Player info found for code %d, will not refresh\n", newPi.Code)
		dbPlaycount := ddr_db.RetrieveLatestPlaycountDetails(db, dbPi.Code)
		if dbPlaycount != nil && playcount.Playcount == dbPlaycount.Playcount {
			glog.Infof("Playcount for %d unchanged. Will not update", dbPi.Code)
			return
		}
	} else {
		glog.Infof("Player info not found for code %d, will refresh\n", newPi.Code)
		refreshDdrUser(client)
	}
	err = ddr_db.AddPlayerDetails(db, *newPi)
	if err != nil {
		glog.Errorf("Failed to update player info for user %s: %s\n", client.GetUsername(), err.Error())
		return
	}

	recentScores, err := ddr.RecentScores(client, newPi.Code)
	if err != nil {
		glog.Errorf("Failed to load recent scores for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
		return
	}
	if recentScores == nil {
		glog.Errorf("failed to load recent scores for code %d\n", newPi.Code)
		err = fmt.Errorf("failed to load recent scores for code %d", newPi.Code)
	}

	workoutData, err := ddr.WorkoutData(client, newPi.Code)
	if err != nil {
		glog.Errorf("Failed to load workout data for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
		return
	}

	recentSongIds := make([]string, 0)
	for _, score := range *recentScores {
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

	dbSongIds := ddr_db.RetrieveSongIds(db)
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
		err = updateNewSongs(client, recentSongIds)
		if err != nil {
			glog.Errorf("Failed to update new songs for client %s\n", client.GetUsername())
			return
		}
	}


	if recentScores != nil {
		err = ddr_db.AddScores(db, *recentScores)
		if err != nil {
			glog.Errorf("Failed to add scores for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
			return
		}

		songsToUpdate := make([]ddr_models.SongDifficulty, 0)

		for _, score := range *recentScores {
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

		statistics, err := ddr.SongStatistics(client, songsToUpdate, newPi.Code)
		if err != nil {
			glog.Errorf("Failed to update song statistics for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
			return
		}
		ddr_db.AddSongStatistics(db, statistics, newPi.Code)
		glog.Infof("Updated song statistics for user %s\n", client.GetUsername())
	}

	if workoutData != nil {
		ddr_db.AddWorkoutData(db, workoutData)
	}

	err = ddr_db.AddPlaycountDetails(db, *playcount)
	if err != nil {
		glog.Errorf("Failed to add playcount details for user %s code %d: %s\n", client.GetUsername(), newPi.Code, err.Error())
		return
	}
	glog.Infof("Profile update complete for user %s\n", client.GetUsername())
	UpdateCookie(client)

	return
}