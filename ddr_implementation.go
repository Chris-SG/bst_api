package main

import (
	"fmt"
	"github.com/chris-sg/eagate/ddr"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/ddr_db"
	"github.com/chris-sg/eagate_models/ddr_models"
	"github.com/chris-sg/eagate_models/user_models"
)

// checkForNewSongs will load the song list from eagate and compare it
// against the song list from the database. Any songs only located in
// eagate will be returned.
func checkForNewSongs(client util.EaClient) (newSongs []string, err error) {
	if !client.LoginState() {
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	siteIds, err := ddr.SongIds(client)
	if err != nil {
		return
	}
	db, err := eagate_db.GetDb()
	if err != nil {
		return
	}
	dbIds := ddr_db.RetrieveSongIds(db)

	for i := len(siteIds)-1; i >= 0; i-- {
		for j, _ := range dbIds {
			if dbIds[j] == siteIds[i] {
				siteIds = append(siteIds[:i], siteIds[i+1:]...)
				dbIds = append(dbIds[:j], dbIds[j+1:]...)
				break
			}
		}
	}
	UpdateCookie(client)
	return
}

// updateNewSongs will load song data and song difficulties for the
// provided songIds slice. This intends to be used after checkForNewSongs
// to update the database.
func updateNewSongs(client util.EaClient, songIds []string) error {
	if !client.LoginState() {
		err := fmt.Errorf("user not logged into eagate")
		return err
	}
	db, _ := eagate_db.GetDb()
	songData, err := ddr.SongData(client, songIds)
	if err != nil {
		return err
	}
	ddr_db.AddSongs(db, songData)
	difficulties, err := ddr.SongDifficulties(client, songIds)
	if err != nil {
		return err
	}
	ddr_db.AddSongDifficulties(db, difficulties)
	UpdateCookie(client)
	return nil
}

// updateSongStatistics will load the client's statistics for the given
// difficulties slice and update the statistics in the database.
func updateSongStatistics(client util.EaClient, difficulties []ddr_models.SongDifficulty) (err error) {
	if !client.LoginState() {
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	pi, _, err := ddr.PlayerInformation(client)
	if err != nil {
		return
	}

	stats, err := ddr.SongStatistics(client, difficulties, pi.Code)
	if err != nil {
		return
	}

	db, _ := eagate_db.GetDb()
	ddr_db.AddPlayerDetails(db, *pi)
	ddr_db.AddSongStatistics(db, stats, pi.Code)
	UpdateCookie(client)
	return
}

// updatePlayerProfile will do a full update of the user's profile. This
// includes updating the player information, the playcount, adding the
// recent scores and updating song statistics.
// TODO: if the user has played more than 50 songs, this will not update
// unknown song statistics. This can currently still be achieved manually.
func updatePlayerProfile(user user_models.User, client util.EaClient) (err error) {
	if !client.LoginState() {
		err = fmt.Errorf("user not logged into eagate")
		return
	}
	db, _ := eagate_db.GetDb()
	newPi, playcount, _ := ddr.PlayerInformation(client)
	newPi.EaGateUser = &user.Name
	dbPi, _ := ddr_db.RetrieveDdrPlayerDetailsByCode(db, newPi.Code)
	if dbPi != nil {
		dbPlaycount := ddr_db.RetrieveLatestPlaycountDetails(db, dbPi.Code)
		if dbPlaycount != nil && playcount.Playcount == dbPlaycount.Playcount {
			return
		}
	} else {
		difficulties := ddr_db.RetrieveValidSongDifficulties(db)
		updateSongStatistics(client, difficulties)
	}
	ddr_db.AddPlayerDetails(db, *newPi)

	recentScores, _ := ddr.RecentScores(client, newPi.Code)
	if recentScores == nil {
		err = fmt.Errorf("failed to load recent scores for code %d", newPi.Code)
	}

	workoutData, _ := ddr.WorkoutData(client, newPi.Code)

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
			return
		}
	}


	if recentScores != nil {
		ddr_db.AddScores(db, *recentScores)

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

		statistics, _ := ddr.SongStatistics(client, songsToUpdate, newPi.Code)
		ddr_db.AddSongStatistics(db, statistics, newPi.Code)
	}

	if workoutData != nil {
		ddr_db.AddWorkoutData(db, workoutData)
	}

	ddr_db.AddPlaycountDetails(db, *playcount)
	UpdateCookie(client)

	return
}