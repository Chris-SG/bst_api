package drs_db

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/models/drs_models"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"strings"
	"time"
)

type DrsDbCommunication interface {
	AddPlayerDetails(details drs_models.PlayerDetails) (errs []error)
	AddPlayerProfileSnapshot(snapshot drs_models.PlayerProfileSnapshot) (errs []error)
	AddSongs(songs []drs_models.Song) (errs []error)
	AddDifficulties(songs []drs_models.Difficulty) (errs []error)
	AddPlayerSongStats(stats []drs_models.PlayerSongStats) (errs []error)
	AddPlayerScores(scores []drs_models.PlayerScore) (errs []error)

	RetrievePlayerDetailsByPlayerCode(code int) (details drs_models.PlayerDetails, errs []error)
	RetrievePlayerDetailsByEaGateUser(eaUser string) (details drs_models.PlayerDetails, errs []error)
	RetrieveRecentPlayerProfileSnapshot(code int) (snapshot drs_models.PlayerProfileSnapshot, errs []error)
	//RetrievePlayerProfileSnapshots(code int, dateFrom time.Time, dateTo time.Time) (snapshots []drs_models.PlayerProfileSnapshot, errs []error)
	//RetrieveSongs() (songs []drs_models.Song, errs []error)
	//RetrieveDifficulties(songs []drs_models.Song) (difficulties []drs_models.Difficulty, errs []error)
	RetrieveSongStatisticsByPlayerCode(code int) (stats []drs_models.PlayerSongStats, errs []error)
	//RetrievePlayerScores(code int) (scores []drs_models.PlayerScore, errs []error)

	RetrieveDataForTable(code int) (json string, errs []error)
}

func CreateDrsDbCommunicationPostgres(db *gorm.DB) DrsDbCommunicationPostgres {
	return DrsDbCommunicationPostgres{db}
}

type DrsDbCommunicationPostgres struct {
	db *gorm.DB
}

const maxBatchSize = 100

func (dbcomm DrsDbCommunicationPostgres) AddPlayerDetails(details drs_models.PlayerDetails) (errs []error) {
	glog.Infof("AddPlayerDetails for %s (code %d)\n", details.EaGateUser, details.Code)
	resultDb := dbcomm.db.Save(&details)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}

	glog.Infof("AddPlayerDetails: %d rows affected\n", resultDb.RowsAffected)
	return
}

func (dbcomm DrsDbCommunicationPostgres) AddPlayerProfileSnapshot(snapshot drs_models.PlayerProfileSnapshot) (errs []error) {
	glog.Infof("AddPlayerProfileSnapshot for code %d\n", snapshot.PlayerCode)
	resultDb := dbcomm.db.Save(&snapshot)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}

	glog.Infof("AddPlayerProfileSnapshot: %d rows affected\n", resultDb.RowsAffected)
	return
}

func (dbcomm DrsDbCommunicationPostgres) AddSongs(songs []drs_models.Song) (errs []error) {
	glog.Infof("AddSongs: %d songs to process\n", len(songs))

	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."drsSongs" VALUES `
	statementEnd := ` ON CONFLICT DO NOTHING;`
	for i := len(songs) - 1; i >= 0; i-- {
		statement = fmt.Sprintf("%s ('%s', '%s', '%s', %d, %d, %d, %d, %d, '%s')",
			statement,
			songs[i].SongId,
			cleanString(songs[i].SongName),
			cleanString(songs[i].ArtistName),
			songs[i].MaxBpm,
			songs[i].MinBpm,
			songs[i].LimitationType,
			songs[i].Genre,
			songs[i].VideoFlags,
			cleanString(songs[i].License))
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

func (dbcomm DrsDbCommunicationPostgres) AddDifficulties(difficulties []drs_models.Difficulty) (errs []error) {
	glog.Infof("AddDifficulties: %d songs to process\n", len(difficulties))

	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."drsDifficulties" VALUES `
	statementEnd := ` ON CONFLICT DO NOTHING;`
	for i := len(difficulties) - 1; i >= 0; i-- {
		statement = fmt.Sprintf("%s ('%s', '%s', %d, '%s')", statement, difficulties[i].Mode, difficulties[i].Difficulty, difficulties[i].Level, difficulties[i].SongId)
		difficulties = difficulties[:len(difficulties)-1]
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
	glog.Infof("AddDifficulties: %d rows affected", totalRowsAffected)
	return nil
}

func (dbcomm DrsDbCommunicationPostgres) AddPlayerSongStats(stats []drs_models.PlayerSongStats) (errs []error) {
	if len(stats) == 0 {
		glog.Infof("AddPlayerSongStats - no statistics to add, aborting")
		return
	}
	glog.Infof("AddPlayerSongStats for playerCode %d (%d statistics)\n", stats[0].PlayerCode, len(stats))
	allSongStatistics, errs := dbcomm.RetrieveSongStatisticsByPlayerCode(stats[0].PlayerCode)
	for i := len(stats) - 1; i >= 0; i-- {
		for _, dbStatistic := range allSongStatistics {
			if stats[i].Equals(dbStatistic) {
				stats = append(stats[:i], stats[i+1:]...)
				break
			}
		}
	}
	if len(stats) == 0 {
		glog.Infof("AddPlayerSongStats - no unique statistics to add, aborting")
		return
	}
	glog.Infof("%d unique statistics for playerCode %d\n", len(stats), stats[0].PlayerCode)

	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."drsPlayerSongStats" VALUES `
	statementEnd := ` ON CONFLICT (song_id, mode, difficulty, player_code) DO UPDATE SET ` +
		`best_score=EXCLUDED.best_score, ` +
		`combo=EXCLUDED.combo, ` +
		`play_count=EXCLUDED.play_count, ` +
		`param=EXCLUDED.param, ` +
		`best_score_time=EXCLUDED.best_score_time, ` +
		`last_play_time=EXCLUDED.last_play_time, ` +
		`p1_code=EXCLUDED.p1_code, ` +
		`p1_score=EXCLUDED.p1_score, ` +
		`p1_perfects=EXCLUDED.p1_perfects, ` +
		`p1_greats=EXCLUDED.p1_greats, ` +
		`p1_goods=EXCLUDED.p1_goods, ` +
		`p1_bads=EXCLUDED.p1_bads, ` +
		`p2_code=EXCLUDED.p2_code, ` +
		`p2_score=EXCLUDED.p2_score, ` +
		`p2_perfects=EXCLUDED.p2_perfects, ` +
		`p2_greats=EXCLUDED.p2_greats, ` +
		`p2_goods=EXCLUDED.p2_goods, ` +
		`p2_bads=EXCLUDED.p2_bads;`

	for i := range stats {
		statement = fmt.Sprintf("%s (%d, %d, %d, %d, '%s', '%s', %d, %d, %d, %d, %d, %d",
			statement,
			stats[i].BestScore,
			stats[i].Combo,
			stats[i].PlayCount,
			stats[i].Param,
			pq.FormatTimestamp(stats[i].BestScoreDateTime),
			pq.FormatTimestamp(stats[i].LastPlayDateTime),
			stats[i].P1Code,
			stats[i].P1Score,
			stats[i].P1Perfects,
			stats[i].P1Greats,
			stats[i].P1Goods,
			stats[i].P1Bads)

		if stats[i].P2Code == nil {
			statement = fmt.Sprintf("%s, NULL, NULL, NULL, NULL, NULL, NULL", statement)
		} else {
			statement = fmt.Sprintf("%s, %d, %d, %d, %d, %d, %d",
				statement,
				*stats[i].P2Code,
				*stats[i].P2Score,
				*stats[i].P2Perfects,
				*stats[i].P2Greats,
				*stats[i].P2Goods,
				*stats[i].P2Bads)
		}

		statement = fmt.Sprintf("%s, %d, '%s', '%s', '%s')",
			statement,
			stats[i].PlayerCode,
			stats[i].SongId,
			stats[i].Mode,
			stats[i].Difficulty)

		batchCount++
		processedCount++
		if batchCount == maxBatchSize || processedCount >= len(stats) {
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
	glog.Infof("AddPlayerSongStats for playerCode %d: %d rows affected\n", stats[0].PlayerCode, totalRowsAffected)
	return nil
}

func (dbcomm DrsDbCommunicationPostgres) AddPlayerScores(scores []drs_models.PlayerScore) (errs []error) {
	glog.Infof("AddPlayerScores with %d scores\n", len(scores))
	batchCount := 0
	processedCount := 0
	statements := make([]string, 0)
	var statement string
	statementBegin := `INSERT INTO public."drsPlayerScores" VALUES `
	statementEnd := ` ON CONFLICT DO NOTHING;`
	for i := range scores {
		statement = fmt.Sprintf("%s ('%s', %d, %d, %d, '%s', %d, %d, %d, %d, %d, %d",
			statement,
			cleanString(scores[i].Shop),
			scores[i].Score,
			scores[i].MaxCombo,
			scores[i].Param,
			pq.FormatTimestamp(scores[i].PlayTime),
			scores[i].P1Code,
			scores[i].P1Score,
			scores[i].P1Perfects,
			scores[i].P1Greats,
			scores[i].P1Goods,
			scores[i].P1Bads)

		if scores[i].P2Code == nil {
			statement = fmt.Sprintf("%s, NULL, NULL, NULL, NULL, NULL, NULL", statement)
		} else {
			statement = fmt.Sprintf("%s, %d, %d, %d, %d, %d, %d",
				statement,
				*scores[i].P2Code,
				*scores[i].P2Score,
				*scores[i].P2Perfects,
				*scores[i].P2Greats,
				*scores[i].P2Goods,
				*scores[i].P2Bads)
		}

		if scores[i].VideoUrl == nil {
			statement = fmt.Sprintf("%s, NULL", statement)
		} else {
			statement = fmt.Sprintf("%s, '%s'", statement, *scores[i].VideoUrl)
		}

		statement = fmt.Sprintf("%s, %d, '%s', '%s', '%s')",
			statement,
			scores[i].PlayerCode,
			scores[i].SongId,
			scores[i].Mode,
			scores[i].Difficulty)

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
		fmt.Println(completeStatement)
		resultDb := dbcomm.db.Exec(completeStatement)
		errors := resultDb.GetErrors()
		if errors != nil && len(errors) != 0 {
			errs = append(errs, errors...)
		}
		totalRowsAffected += resultDb.RowsAffected
	}
	glog.Infof("AddPlayerScores: %d rows affected\n", totalRowsAffected)

	return
}


func (dbcomm DrsDbCommunicationPostgres) RetrievePlayerDetailsByPlayerCode(code int) (details drs_models.PlayerDetails, errs []error) {
	glog.Infof("Retrieve player details for code %d\n", code)
	resultDb := dbcomm.db.Model(&drs_models.PlayerDetails{}).Where("player_code = ?", code).First(&details)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DrsDbCommunicationPostgres) RetrievePlayerDetailsByEaGateUser(eaUser string) (details drs_models.PlayerDetails, errs []error) {
	glog.Infof("Retrieve player details for eauser %s\n", eaUser)
	resultDb := dbcomm.db.Model(&drs_models.PlayerDetails{}).Where("eagate_user = ?", eaUser).First(&details)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DrsDbCommunicationPostgres) RetrieveRecentPlayerProfileSnapshot(code int) (snapshot drs_models.PlayerProfileSnapshot, errs []error) {
	glog.Infof("Retrieve recent snapshot for code %d\n", code)
	resultDb := dbcomm.db.Model(&drs_models.PlayerProfileSnapshot{}).Where("player_code = ?", code).Order("play_count desc").First(&snapshot)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm DrsDbCommunicationPostgres) RetrieveSongStatisticsByPlayerCode(code int) (stats []drs_models.PlayerSongStats, errs []error) {
	glog.Infof("RetrieveSongStatisticsByPlayerCode for player code %d\n", code)
	resultDb := dbcomm.db.Model(&drs_models.PlayerSongStats{}).Where("player_code = ?", code).Scan(&stats)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

type DrsDataTable struct {
	Title string `json:"title"`
	Artist string `json:"artist"`
	Mode string `json:"mode"`
	Difficulty string `json:"difficulty"`
	Level string `json:"level"`

	Score int `json:"score"`
	PlayCount int `json:"playcount" gorm:"column:playcount"`
	BestScoreDateTime time.Time `json:"bestscoretime" gorm:"column:bestscoredatetime"`

	P1Code int `json:"p1code" gorm:"column:p1code"`
	P1Perfects int `json:"p1perfects" gorm:"column:p1perfects"`
	P1Greats int `json:"p1greats" gorm:"column:p1greats"`
	P1Goods int `json:"p1goods" gorm:"column:p1goods"`
	P1Bads int `json:"p1bads" gorm:"column:p1bads"`

	P2Code int `json:"p2code" gorm:"column:p2code"`
	P2Perfects int `json:"p2perfects" gorm:"column:p2perfects"`
	P2Greats int `json:"p2greats" gorm:"column:p2greats"`
	P2Goods int `json:"p2goods" gorm:"column:p2goods"`
	P2Bads int `json:"p2bads" gorm:"column:p2bads"`

	SongId string `json:"id" gorm:"column:id"`
	Code int `json:"code"`
	Param int `json:"param"`

	//Combo
	//LastPlayDateTime
}

func (dbcomm DrsDbCommunicationPostgres) RetrieveDataForTable(code int) (resultJson string, errs []error) {
	stats := make([]DrsDataTable, 0)

	resultDb := dbcomm.db.
		Table("public.\"drsDifficulties\" diff").
		Select("diff.level as level," +
			"diff.mode as mode," +
			"diff.difficulty as difficulty," +
			"diff.level as level," +
			"song.name as title," +
			"song.artist as artist," +
			"stat.best_score as score," +
			"stat.play_count as playcount," +
			"stat.best_score_time as bestscoredatetime," +
			"stat.p1_code as p1code," +
			"stat.p1_perfects as p1perfects," +
			"stat.p1_greats as p1greats," +
			"stat.p1_goods as p1goods," +
			"stat.p1_bads as p1bads," +
			"stat.p2_code as p2code," +
			"stat.p2_perfects as p2perfects," +
			"stat.p2_greats as p2greats," +
			"stat.p2_goods as p2goods," +
			"stat.p2_bads as p2bads," +
			"diff.song_id as id," +
			"stat.player_code as code," +
			"stat.param as param").
		Joins("inner join public.\"drsSongs\" song on diff.song_id = song.song_id").
		Joins("left outer join public.\"drsPlayerSongStats\" stat on " +
			"diff.song_id = stat.song_id AND " +
			"diff.mode = stat.mode AND " +
			"diff.difficulty = stat.difficulty AND " +
			"stat.player_code = ?", code).
		Order("diff.mode desc, diff.level").
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

	resultJson = string(result)
	return
}


func cleanString(in string) string {
	return strings.ReplaceAll(in, "'", "&#39;")
}

func fixString(in string) string {
	return strings.ReplaceAll(in, "&#39;", "'")
	return strings.ReplaceAll(in, "&amp;", "&")
}