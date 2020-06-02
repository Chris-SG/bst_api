package ddr_db

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)

type DdrDbCommunication interface {
	AddSongs(songs []ddr_models.Song) (errs []error)
	RetrieveSongIds() (songIds []string, errs []error)
	RetrieveSongsById(songIds []string, ordering []string) (songs []ddr_models.Song, errs []error)
	RetrieveJacketForSongId(songId string) (jacket string, errs []error)
	RetrieveJacketsForSongIds(songIds []string) (jackets map[string] string, errs []error)

	AddDifficulties(difficulties []ddr_models.SongDifficulty) (errs []error)
	RetrieveDifficulties() (difficulties []ddr_models.SongDifficulty, errs []error)
	RetrieveValidDifficulties() (difficulties []ddr_models.SongDifficulty, errs []error)
	RetrieveDifficultiesById(songIds []string) (difficulties []ddr_models.SongDifficulty, errs []error)
	RetrieveValidDifficultiesById(songIds []string) (difficulties []ddr_models.SongDifficulty, errs []error)

	AddPlayerDetails(details ddr_models.PlayerDetails) (errs []error)
	RetrievePlayerDetailsByEaGateUser(eaGateUser string) (details ddr_models.PlayerDetails, exists bool, errs []error)
	RetrievePlayerDetailsByPlayerCode(code int) (details ddr_models.PlayerDetails, errs []error)

	AddPlaycounts(playcountDetails []ddr_models.Playcount) (errs []error)
	RetrievePlaycountsByPlayerCode(code int) (playcounts []ddr_models.Playcount, errs []error)
	RetrieveLatestPlaycountByPlayerCode(code int) (playcount ddr_models.Playcount, errs []error)
	RetrievePlaycountsByPlayerCodeInDateRange(code int, startDate time.Time, endDate time.Time) (playcounts []ddr_models.Playcount, errs []error)

	AddSongStatistics(statistics []ddr_models.SongStatistics) (errs []error)
	RetrieveSongStatisticsByPlayerCode(code int, songIds []string) (statistics []ddr_models.SongStatistics, errs []error)

	AddScores(scores []ddr_models.Score) (errs []error)
	RetrieveScoresByPlayerCode(code int) (scores []ddr_models.Score, errs []error)
	RetrieveSongScores(code int, songId string, mode string, difficulty string, ordering []string) (scores []ddr_models.Score, errs []error)

	AddWorkoutData(workoutData []ddr_models.WorkoutData) (errs []error)
	RetrieveWorkoutDataByPlayerCode(code int) (workoutData []ddr_models.WorkoutData, errs []error)
	RetrieveWorkoutDataByPlayerCodeInDateRange(code int, startDate time.Time, endDate time.Time) (workoutData []ddr_models.WorkoutData, errs []error)

	RetrieveExtendedScoreStatisticsByPlayerCode(code int) (statisticsJson string, errs []error)
}

func CreateDdrDbCommunicationPostgres(db *gorm.DB) DdrDbCommunicationPostgres {
	return DdrDbCommunicationPostgres{db}
}

type DdrDbCommunicationPostgres struct {
	db *gorm.DB
}

const maxBatchSize = 100

func (dbcomm DdrDbCommunicationPostgres) AddSongs(songs []ddr_models.Song) (errs []error) {
	glog.Infof("AddSongs: %d songs to process\n", len(songs))
	currentIds, errs := dbcomm.RetrieveSongIds()
	for i := len(songs)-1; i >= 0; i-- {
		for _, id := range currentIds {
			if id == songs[i].Id {
				songs = append(songs[:i], songs[i+1:]...)
				break
			}
		}
	}
	glog.Infof("AddSongs: %d new songs\n", len(songs))

	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."ddrSongs" VALUES `
	statementEnd := ` ON CONFLICT DO NOTHING;`
	for i := len(songs)-1; i >= 0; i-- {
		statement = fmt.Sprintf("%s ('%s', '%s', '%s', '%s')", statement, songs[i].Id, cleanString(songs[i].Name), cleanString(songs[i].Artist), songs[i].Image)
		songs = songs[:len(songs)-1]
		batchCount++
		processedCount++
		if batchCount == maxBatchSize || i == 0 {
			statement = fmt.Sprintf("%s%s%s", statementBegin, statement, statementEnd)
			statements = append(statements, statement)
			statement = ""
		} else {
			statement = fmt.Sprintf("%s,", statement)
		}
	}

	totalRowsAffected := int64(0)
	for _, completeStatement := range statements {
		resultDb := dbcomm.db.Exec(completeStatement)
		errors := resultDb.GetErrors()
		if errors != nil && len(errors) != 0 {
			errs = append(errs, errors...)
		}
		totalRowsAffected += resultDb.RowsAffected
	}
	glog.Infof("AddSongs: %d rows affected", totalRowsAffected)
	return nil
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveSongIds() (songIds []string, errs []error) {
	glog.Infoln("RetrieveSongIds")
	resultDb := dbcomm.db.Model(&ddr_models.Song{}).Select("id").Pluck("id", &songIds)
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveSongsById(songIds []string, ordering []string) (songs []ddr_models.Song, errs []error) {
	glog.Infof("RetrieveSongsByIds for %d ids\n", len(songIds))
	resultDb := dbcomm.db.Model(&ddr_models.Song{}).Select([]string{"id", "name", "artist"}).Where("id IN (?)", songIds)
	for _, order := range ordering {
		resultDb = resultDb.Order(order)
	}
	resultDb = resultDb.Scan(&songs)
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveJacketForSongId(songId string) (jacket string, errs []error) {
	glog.Infof("getting for id %s\n", songId)

	jacketSlice := make([]string, 0)
	resultDb := dbcomm.db.Model(&ddr_models.Song{}).Limit(1).Select("image").Where("id = ?", songId).Pluck("image", &jacketSlice)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}

	if len(jacketSlice) > 0 {
		jacket = jacketSlice[0]
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveJacketsForSongIds(songIds []string) (jackets map[string] string, errs []error) {
	glog.Infof("RetrieveJacketsForSongIds for %d ids\n", len(songIds))
	jackets = make(map[string]string)

	type tmp struct {
		id string
		image string
	}

	data := make([]tmp, 0)

	resultDb := dbcomm.db.Model(&ddr_models.Song{}).Select([]string{"id", "image"}).Where("id IN (?)", songIds).Scan(&data)
	for _, v := range data {
		jackets[v.id] = v.image
	}

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}


func (dbcomm DdrDbCommunicationPostgres) AddDifficulties(difficulties []ddr_models.SongDifficulty) (errs []error) {
	glog.Infof("AddSongDifficulties for %d difficulties\n", len(difficulties))
	allSongDifficulties, errs := dbcomm.RetrieveDifficulties()
	for i := len(difficulties)-1; i >= 0; i-- {
		for _, dbDifficulty := range allSongDifficulties {
			if  difficulties[i].SongId == dbDifficulty.SongId &&
				difficulties[i].Mode == dbDifficulty.Mode &&
				difficulties[i].Difficulty == dbDifficulty.Difficulty &&
				difficulties[i].DifficultyValue == dbDifficulty.DifficultyValue {
					difficulties = append(difficulties[:i], difficulties[i+1:]...)
				break
			}
		}
	}
	glog.Infof("AddSongDifficulties for %d new or updated difficulties\n", len(difficulties))
	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."ddrSongDifficulties" VALUES `
	statementEnd := ` ON CONFLICT (song_id, mode, difficulty) DO UPDATE SET difficulty_value=EXCLUDED.difficulty_value;`
	for i := range difficulties {
		statement = fmt.Sprintf("%s ('%s', '%s', '%s', %d)",
			statement,
			difficulties[i].SongId,
			difficulties[i].Mode,
			difficulties[i].Difficulty,
			difficulties[i].DifficultyValue)

		batchCount++
		processedCount++
		if batchCount == maxBatchSize || processedCount >= len(difficulties) {
			statement = fmt.Sprintf("%s%s%s", statementBegin, statement, statementEnd)
			statements = append(statements, statement)
			statement = ""
		} else {
			statement = fmt.Sprintf("%s,", statement)
		}
	}

	totalRowsAffected := int64(0)
	for _, completeStatement := range statements {
		resultDb := dbcomm.db.Exec(completeStatement)
		errors := resultDb.GetErrors()
		if errors != nil && len(errors) != 0 {
			errs = append(errs, errors...)
		}
		totalRowsAffected += resultDb.RowsAffected
	}
	glog.Infof("AddSongDifficulties: %d rows affected\n", totalRowsAffected)
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveDifficulties() (difficulties []ddr_models.SongDifficulty, errs []error) {
	glog.Infoln("RetrieveAllSongDifficulties")
	resultDb := dbcomm.db.Model(&ddr_models.SongDifficulty{}).Scan(&difficulties)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveValidDifficulties() (difficulties []ddr_models.SongDifficulty, errs []error) {
	glog.Infoln("RetrieveValidSongDifficulties")
	resultDb := dbcomm.db.Model(&ddr_models.SongDifficulty{}).Where("difficulty_value > -1").Scan(&difficulties)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveDifficultiesById(songIds []string) (difficulties []ddr_models.SongDifficulty, errs []error) {
	glog.Infoln("RetrieveSongDifficultiesById")
	resultDb := dbcomm.db.Model(&ddr_models.SongDifficulty{}).Where("song_id IN (?)", songIds).Scan(&difficulties)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveValidDifficultiesById(songIds []string) (difficulties []ddr_models.SongDifficulty, errs []error) {
	glog.Infoln("RetrieveValidDifficultiesById")
	resultDb := dbcomm.db.Model(&ddr_models.SongDifficulty{}).Where("song_id IN (?) AND difficulty_value > -1", songIds).Scan(&difficulties)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) AddPlayerDetails(details ddr_models.PlayerDetails) (errs []error) {
	glog.Infof("AddPlayerDetails for %s (code %d)\n", details.EaGateUser, details.Code)
	resultDb := dbcomm.db.Save(&details)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}

	glog.Infof("AddPlayerDetails: %d rows affected\n", resultDb.RowsAffected)
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrievePlayerDetailsByEaGateUser(eaGateUser string) (details ddr_models.PlayerDetails, exists bool, errs []error) {
	glog.Infof("RetrieveDdrPlayerDetailsByEaGateUser for eaUser %s\n", eaGateUser)
	eaGateUser = strings.ToLower(eaGateUser)
	resultDb := dbcomm.db.Model(&ddr_models.PlayerDetails{}).Where("eagate_user = ?", eaGateUser).First(&details)
	if resultDb.RecordNotFound() {
		exists = false
		return
	}
	exists = true
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrievePlayerDetailsByPlayerCode(code int) (details ddr_models.PlayerDetails, errs []error) {
	resultDb := dbcomm.db.Model(&ddr_models.PlayerDetails{}).Where("code = ?", code).First(&details)
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) AddPlaycounts(playcountDetails []ddr_models.Playcount) (errs []error) {
	glog.Infof("AddPlaycounts adding %d datapoints\n", len(playcountDetails))

	processedCount := 0
	var statement string
	statementBegin := `INSERT INTO public."ddrPlaycount" VALUES `
	statementEnd := ` ON CONFLICT (last_play_date, player_code) do nothing;`
	for _, playcount := range playcountDetails {
		statement = fmt.Sprintf("%s (%d, '%s', %d, '%s', %d, '%s', %d)",
			statement,
			playcount.Playcount,
			pq.FormatTimestamp(playcount.LastPlayDate),
			playcount.SinglePlaycount,
			pq.FormatTimestamp(playcount.SingleLastPlayDate),
			playcount.DoublePlaycount,
			pq.FormatTimestamp(playcount.DoubleLastPlayDate),
			playcount.PlayerCode)

		processedCount++
		if processedCount >= len(playcountDetails) {
			statement = fmt.Sprintf("%s%s%s", statementBegin, statement, statementEnd)
		} else {
			statement = fmt.Sprintf("%s,", statement)
		}
	}

	resultDb := dbcomm.db.Exec(statement)
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	glog.Infof("AddPlaycounts: %d rows affected\n", resultDb.RowsAffected)
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrievePlaycountsByPlayerCode(code int) (playcounts []ddr_models.Playcount, errs []error) {
	glog.Infof("RetrievePlaycountsByPlayerCode for playerCode %d\n", code)
	resultDb := dbcomm.db.Model(&ddr_models.Playcount{}).Where("player_code = ?", code).Scan(&playcounts)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveLatestPlaycountByPlayerCode(code int) (playcount ddr_models.Playcount, errs []error) {
	glog.Infof("RetrieveLatestPlaycountByPlayerCode for playerCode %d\n", code)
	resultDb := dbcomm.db.Model(&ddr_models.Playcount{}).Where("player_code = ?", code).Order("playcount DESC", true).First(&playcount)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrievePlaycountsByPlayerCodeInDateRange(code int, startDate time.Time, endDate time.Time) (playcounts []ddr_models.Playcount, errs []error) {
	glog.Infof("RetrievePlaycountsByPlayerCodeInDateRange for playerCode %d range %d-%d\n", code, startDate.String(), endDate.String())
	resultDb := dbcomm.db.Model(&ddr_models.Playcount{}).Where("player_code = ?", code).
		Where("last_play_date between ? and ?", pq.FormatTimestamp(startDate), pq.FormatTimestamp(endDate)).
		Scan(&playcounts)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) AddSongStatistics(statistics []ddr_models.SongStatistics) (errs []error) {
	if len(statistics) == 0 {
		glog.Infof("AddSongStatistics - no statistics to add, aborting")
		return
	}
	glog.Infof("AddSongStatistics for playerCode %d (%d statistics)\n", statistics[0].PlayerCode, len(statistics))
	allSongStatistics, errs := dbcomm.RetrieveSongStatisticsByPlayerCode(statistics[0].PlayerCode, []string{})
	for i := len(statistics)-1; i >= 0; i-- {
		for _, dbStatistic := range allSongStatistics {
			if statistics[i].Equals(dbStatistic) {
				statistics = append(statistics[:i], statistics[i+1:]...)
				break
			}
		}
	}
	if len(statistics) == 0 {
		glog.Infof("AddSongStatistics - no unique statistics to add, aborting")
		return
	}
	glog.Infof("%d unique statistics for playerCode %d\n", len(statistics), statistics[0].PlayerCode)

	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."ddrSongStatistics" VALUES `
	statementEnd := ` ON CONFLICT (song_id, mode, difficulty, player_code) DO UPDATE SET ` +
		`score_record=EXCLUDED.score_record, ` +
		`clear_lamp=EXCLUDED.clear_lamp, ` +
		`rank=EXCLUDED.rank, ` +
		`playcount=EXCLUDED.playcount, ` +
		`clearcount=EXCLUDED.clearcount, ` +
		`maxcombo=EXCLUDED.maxcombo, ` +
		`lastplayed=EXCLUDED.lastplayed;`
	for i := range statistics {
		statement = fmt.Sprintf("%s (%d, '%s', '%s', %d, %d, %d, '%s', '%s', '%s', '%s', %d)",
			statement,
			statistics[i].BestScore,
			statistics[i].Lamp,
			statistics[i].Rank,
			statistics[i].PlayCount,
			statistics[i].ClearCount,
			statistics[i].MaxCombo,
			pq.FormatTimestamp(statistics[i].LastPlayed),
			statistics[i].SongId,
			statistics[i].Mode,
			statistics[i].Difficulty,
			statistics[i].PlayerCode)

		batchCount++
		processedCount++
		if batchCount == maxBatchSize || processedCount >= len(statistics) {
			statement = fmt.Sprintf("%s%s%s", statementBegin, statement, statementEnd)
			statements = append(statements, statement)
			statement = ""
		} else {
			statement = fmt.Sprintf("%s,", statement)
		}
	}

	totalRowsAffected := int64(0)
	for _, completeStatement := range statements {
		resultDb := dbcomm.db.Exec(completeStatement)
		errors := resultDb.GetErrors()
		if errors != nil && len(errors) != 0 {
			errs = append(errs, errors...)
		}
		totalRowsAffected += resultDb.RowsAffected
	}
	glog.Infof("AddSongStatistics for playerCode %d: %d rows affected\n", statistics[0].PlayerCode, totalRowsAffected)
	return nil

}

func (dbcomm DdrDbCommunicationPostgres) RetrieveSongStatisticsByPlayerCode(code int, songIds []string) (statistics []ddr_models.SongStatistics, errs []error) {
	glog.Info("RetrieveSongStatisticsByPlayerCode for player code %d\n", code)
	resultDb := dbcomm.db.Model(&ddr_models.SongStatistics{}).Where("player_code = ?", code)
	if len(songIds) > 0 {
		resultDb = resultDb.Where("song_id IN (?)", songIds)
	}
	resultDb = resultDb.Scan(&statistics)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) AddScores(scores []ddr_models.Score) (errs []error) {
	glog.Info("AddScores with %d scores\n", len(scores))
	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."ddrScores" VALUES `
	statementEnd := ` ON CONFLICT DO NOTHING;`
	for i := range scores {
		statement = fmt.Sprintf("%s (%d, '%s', '%s', '%s', '%s', '%s', %d)",
			statement,
			scores[i].Score,
			strconv.FormatBool(scores[i].ClearStatus),
			pq.FormatTimestamp(scores[i].TimePlayed),
			scores[i].SongId,
			scores[i].Mode,
			scores[i].Difficulty,
			scores[i].PlayerCode)

		batchCount++
		processedCount++
		if batchCount == maxBatchSize || processedCount >= len(scores) {
			statement = fmt.Sprintf("%s%s%s", statementBegin, statement, statementEnd)
			statements = append(statements, statement)
			statement = ""
		} else {
			statement = fmt.Sprintf("%s,", statement)
		}
	}

	totalRowsAffected := int64(0)
	for _, completeStatement := range statements {
		resultDb := dbcomm.db.Exec(completeStatement)
		errors := resultDb.GetErrors()
		if errors != nil && len(errors) != 0 {
			errs = append(errs, errors...)
		}
		totalRowsAffected += resultDb.RowsAffected
	}
	glog.Infof("AddScores: %d rows affected\n", totalRowsAffected)

	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveScoresByPlayerCode(code int) (scores []ddr_models.Score, errs []error) {
	glog.Infof("RetrieveScoresByPlayerCode for player code %d\n", code)
	resultDb := dbcomm.db.Model(&ddr_models.Score{}).Where("player_code = ?", code).Scan(&scores)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveSongScores(code int, songId string, mode string, difficulty string, ordering []string) (scores []ddr_models.Score, errs []error) {
	chain := dbcomm.db.Model(&ddr_models.Score{})
	if code == 0 {
		errs = append(errs, fmt.Errorf("no user code specified"))
		return
	}
	if songId == "" {
		errs = append(errs, fmt.Errorf("no song id specified"))
		return
	}
	chain = chain.Where("player_code = ? AND song_id = ?", fmt.Sprintf("%d", code), songId)
	if mode != "" {
		chain = chain.Where("mode = ?", strings.ToUpper(mode))
	}
	if difficulty != "" {
		chain = chain.Where("difficulty = ?", strings.ToUpper(difficulty))
	}
	for _, order := range ordering {
		chain = chain.Order(order)
	}

	resultDb := chain.Find(&scores)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	glog.Infof("SongScores: %d rows\n", len(scores))

	return
}

func (dbcomm DdrDbCommunicationPostgres) AddWorkoutData(workoutData []ddr_models.WorkoutData) (errs []error) {
	glog.Infof("AddWorkoutData: %d data points\n", len(workoutData))
	processedCount := 0
	var statement string
	statementBegin := `INSERT INTO public."ddrWorkoutData" VALUES `
	statementEnd := ` ON CONFLICT (date, player_code) DO UPDATE SET playcount=EXCLUDED.playcount, kcal=EXCLUDED.kcal;`
	for i := range workoutData {
		statement = fmt.Sprintf("%s ('%s', '%d', '%f', %d)",
			statement,
			pq.FormatTimestamp(workoutData[i].Date),
			workoutData[i].PlayCount,
			workoutData[i].Kcal,
			workoutData[i].PlayerCode)

		processedCount++
		if processedCount >= len(workoutData) {
			statement = fmt.Sprintf("%s%s%s", statementBegin, statement, statementEnd)
		} else {
			statement = fmt.Sprintf("%s,", statement)
		}
	}

	resultDb := dbcomm.db.Exec(statement)
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	glog.Infof("AddWorkoutData: %d rows affected\n", resultDb.RowsAffected)
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveWorkoutDataByPlayerCode(code int) (workoutData []ddr_models.WorkoutData, errs []error) {
	glog.Infof("RetrieveWorkoutDataByPlayerCode for player code %d\n", code)
	resultDb := dbcomm.db.Model(&ddr_models.WorkoutData{}).Where("player_code = ?", code).Scan(&workoutData)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	glog.Infof("RetrieveWorkoutDataByPlayerCode for player code %d: %d data points\n", code, len(workoutData))
	return
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveWorkoutDataByPlayerCodeInDateRange(code int, startDate time.Time, endDate time.Time) (workoutData []ddr_models.WorkoutData, errs []error) {
	glog.Infof("player code %d days ago %s to %s\n", code, startDate.String(), endDate.String())
	resultDb := dbcomm.db.
		Model(&ddr_models.WorkoutData{}).
		Where("player_code = ?", code).
		Where("date between ? and ?", startDate, endDate).
		Scan(&workoutData)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	glog.Infof("player code %d: %d data points\n", code, len(workoutData))
	return
}
type DdrStatisticsTable struct {
	Level int `json:"level"`
	Title string `json:"title"`
	Artist string `json:"artist"`
	Mode string `json:"mode"`
	Difficulty string `json:"difficulty"`
	Lamp string `json:"lamp"`
	Rank string `json:"rank"`
	Score int `json:"score"`
	PlayCount int `gorm:"column:playcount" json:"playcount"`
	ClearCount int `gorm:"column:clearcount" json:"clearcount"`
	MaxCombo int `gorm:"column:maxcombo" json:"maxcombo"`
	Id string `gorm:"column:id" json:"id"`
}

func (dbcomm DdrDbCommunicationPostgres) RetrieveExtendedScoreStatisticsByPlayerCode(code int) (statisticsJson string, errs []error) {
	stats := make([]DdrStatisticsTable, 0)

	resultDb := dbcomm.db.
		Table("public.\"ddrSongDifficulties\" diff").
		Select("diff.difficulty_value as level," +
			"diff.mode as mode," +
			"diff.difficulty as difficulty," +
			"song.name as title," +
			"song.artist as artist," +
			"stat.clear_lamp as lamp," +
			"stat.rank as rank," +
			"stat.score_record as score," +
			"stat.playcount as playcount," +
			"stat.clearcount as clearcount," +
			"stat.maxcombo as maxcombo," +
			"diff.song_id as id").
		Joins("inner join public.\"ddrSongs\" song on diff.song_id = song.id").
		Joins("left outer join public.\"ddrSongStatistics\" stat on " +
			"diff.song_id = stat.song_id AND " +
			"diff.mode = stat.mode AND " +
			"diff.difficulty = stat.difficulty AND " +
			"stat.player_code = ?", code).
		Where("diff.difficulty_value != -1").
		Order("diff.mode desc, diff.difficulty_value").
		Scan(&stats)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}

	for i := range stats {
		stats[i].Title = fixString(stats[i].Title)
		stats[i].Artist = fixString(stats[i].Artist)
	}

	result, err := json.Marshal(stats)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to convert loaded extended statistics to json for code %d: %s", code, err.Error()))
		return
	}

	statisticsJson = string(result)
	return
}

func cleanString(in string) string {
	return strings.ReplaceAll(in, "'", "&#39;")
}

func fixString(in string) string {
	return strings.ReplaceAll(in, "&#39;", "'")
	return strings.ReplaceAll(in, "&amp;", "&")
}