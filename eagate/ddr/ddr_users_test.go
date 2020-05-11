package ddr

import (
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/golang/glog"
	"testing"
	"time"
)

func TestPlayerInformationFromDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/player/index.html"

	// Setup expected results

	expectedPlayerInformation := ddr_models.PlayerDetails{
		Code:        12345678,
		Name:        "EAGATE",
		Prefecture:  "オーストラリア",
		SingleRank:  "段位なし",
		DoubleRank:  "段位なし",
		Affiliation: "所属なし",
	}

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	playerInformation, err := playerInformationFromPlayerDocument(document)
	if err != nil {
		t.Fatalf("error in playerInformationFromPlayerDocument: %s", err.Error())
	}

	if  playerInformation.Code != expectedPlayerInformation.Code ||
		playerInformation.Name != expectedPlayerInformation.Name ||
		playerInformation.Prefecture != expectedPlayerInformation.Prefecture ||
		playerInformation.SingleRank != expectedPlayerInformation.SingleRank ||
		playerInformation.DoubleRank != expectedPlayerInformation.DoubleRank ||
		playerInformation.Affiliation != expectedPlayerInformation.Affiliation {
		t.Errorf("player information did not match, expected %+#v but got %+#v", expectedPlayerInformation, playerInformation)
	}
}

func TestPlaycountFromDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/player/index.html"

	// Setup expected results

	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	expectedLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedSingleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedDoubleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-02-20 19:11:10", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}

	expectedPlaycount := ddr_models.Playcount{
		Playcount:          380,
		LastPlayDate:       expectedLastPlayDate,
		SinglePlaycount:    360,
		SingleLastPlayDate: expectedSingleLastPlayDate,
		DoublePlaycount:    20,
		DoubleLastPlayDate: expectedDoubleLastPlayDate,
		PlayerCode:         12345678,
	}

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	playcount, err := playcountFromPlayerDocument(document)
	if err != nil {
		t.Fatalf("error in playerInformationFromPlayerDocument: %s", err.Error())
	}

	if playcount.Playcount != expectedPlaycount.Playcount ||
		playcount.LastPlayDate.String() != expectedPlaycount.LastPlayDate.String() ||
		playcount.SinglePlaycount != expectedPlaycount.SinglePlaycount ||
		playcount.SingleLastPlayDate.String() != expectedPlaycount.SingleLastPlayDate.String() ||
		playcount.DoublePlaycount != expectedPlaycount.DoublePlaycount ||
		playcount.DoubleLastPlayDate.String() != expectedPlaycount.DoubleLastPlayDate.String() ||
		playcount.PlayerCode != expectedPlaycount.PlayerCode {
		t.Errorf("playcount information did not match, expected %+#v but got %+#v", expectedPlaycount, playcount)
	}
}

func TestPlayerInformationForClient(t *testing.T) {
	// Setup test
	const playerProfilePage = "./test_data/player/index.html"
	const playerProfileUri = "https://p.eagate.573.jp/game/ddr/ddra20/p/playdata/index.html"
	uriMapping := make(map[string]string)
	uriMapping[playerProfileUri] = playerProfilePage

	c, s := testServerAndClient(uriMapping)
	defer s.Close()

	// Setup expected results

	expectedPlayerInformation := ddr_models.PlayerDetails{
		Code:        12345678,
		Name:        "EAGATE",
		Prefecture:  "オーストラリア",
		SingleRank:  "段位なし",
		DoubleRank:  "段位なし",
		Affiliation: "所属なし",
	}

	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	expectedLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedSingleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-03-18 18:52:59", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}
	expectedDoubleLastPlayDate, err := time.ParseInLocation(timeFormat, "2020-02-20 19:11:10", timeLocation)
	if err != nil {
		t.Fatalf("error in time parsing for expectedLastPlayDate: %s", err.Error())
	}

	expectedPlaycount := ddr_models.Playcount{
		Playcount:          380,
		LastPlayDate:       expectedLastPlayDate,
		SinglePlaycount:    360,
		SingleLastPlayDate: expectedSingleLastPlayDate,
		DoublePlaycount:    20,
		DoubleLastPlayDate: expectedDoubleLastPlayDate,
		PlayerCode:         12345678,
	}

	// Run test
	playerInformation, playcount, err := PlayerInformationForClient(c)

	if  playerInformation.Code != expectedPlayerInformation.Code ||
		playerInformation.Name != expectedPlayerInformation.Name ||
		playerInformation.Prefecture != expectedPlayerInformation.Prefecture ||
		playerInformation.SingleRank != expectedPlayerInformation.SingleRank ||
		playerInformation.DoubleRank != expectedPlayerInformation.DoubleRank ||
		playerInformation.Affiliation != expectedPlayerInformation.Affiliation {
		t.Errorf("player information did not match, expected %+#v but got %+#v", expectedPlayerInformation, playerInformation)
	}

	if playcount.Playcount != expectedPlaycount.Playcount ||
		playcount.LastPlayDate.String() != expectedPlaycount.LastPlayDate.String() ||
		playcount.SinglePlaycount != expectedPlaycount.SinglePlaycount ||
		playcount.SingleLastPlayDate.String() != expectedPlaycount.SingleLastPlayDate.String() ||
		playcount.DoublePlaycount != expectedPlaycount.DoublePlaycount ||
		playcount.DoubleLastPlayDate.String() != expectedPlaycount.DoubleLastPlayDate.String() ||
		playcount.PlayerCode != expectedPlaycount.PlayerCode {
		t.Errorf("playcount information did not match, expected %+#v but got %+#v", expectedPlaycount, playcount)
	}
}

func TestChartStatisticsFromDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/music_detail/1PoOQPd0D01Q9O0doiQQQ8D8Q096bDq9.html"
	const testId = "1PoOQPd0D01Q9O0doiQQQ8D8Q096bDq9"

	// Setup expected results
	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	difficulty := ddr_models.SongDifficulty{
		SongId:          "1PoOQPd0D01Q9O0doiQQQ8D8Q096bDq9",
		Mode:            "SINGLE",
		Difficulty:      "BEGINNER",
		DifficultyValue: 3,
	}

	expectedTime, _ := time.ParseInLocation(timeFormat, "2018-06-07 19:11:17", timeLocation)

	expectedStatistics := ddr_models.SongStatistics{
		BestScore:  831790,
		Lamp:       "---",
		Rank:       "A",
		PlayCount:  1,
		ClearCount: 1,
		MaxCombo:   108,
		LastPlayed: expectedTime,
		SongId:     difficulty.SongId,
		Mode:       difficulty.Mode,
		Difficulty: difficulty.Difficulty,
		PlayerCode: 12345678,
	}

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	statistics, err := chartStatisticsFromDocument(document, 12345678, difficulty)

	if err != nil {
		t.Errorf("failed to load chart stats from document: %s", err.Error())
	}

	if statistics.BestScore != expectedStatistics.BestScore ||
		statistics.Lamp != expectedStatistics.Lamp ||
		statistics.Rank != expectedStatistics.Rank ||
		statistics.PlayCount != expectedStatistics.PlayCount ||
		statistics.ClearCount != expectedStatistics.ClearCount ||
		statistics.MaxCombo != expectedStatistics.MaxCombo ||
		statistics.LastPlayed.String() != expectedStatistics.LastPlayed.String() ||
		statistics.SongId != expectedStatistics.SongId ||
		statistics.Mode != expectedStatistics.Mode ||
		statistics.Difficulty != expectedStatistics.Difficulty ||
		statistics.PlayerCode != expectedStatistics.PlayerCode {
		t.Errorf("statistics do not match, expected %+#v but got %+#v", expectedStatistics, statistics)
	}
}

func TestNoPlaySongStatisticsForDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/music_detail/8bQQ0lP96186D8Ibo8IoOd6o16qioiIo.html"
	const testId = "8bQQ0lP96186D8Ibo8IoOd6o16qioiIo"

	// Setup expected results

	expectedStatistics := ddr_models.SongStatistics{}

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	statistics, err := chartStatisticsFromDocument(document, 12345678, ddr_models.SongDifficulty{})

	if err != nil {
		t.Errorf("failed to load chart stats from document: %s", err.Error())
	}

	if statistics != expectedStatistics {
		t.Errorf("statistics do not match, expected %+#v but got %+#v", expectedStatistics, statistics)
	}
}

func TestSongStatisticsForClient(t *testing.T) {

}

func TestRecentScoresFromDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/player/recent_scores.html"

	// Setup expected results
	timeFormat := "2006-01-02 15:04:05"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	expectedRecentScores := make([]ddr_models.Score, 0)

	date, _ := time.ParseInLocation(timeFormat, "2020-03-18 18:52:17", timeLocation)
	expectedRecentScores = append(expectedRecentScores, ddr_models.Score{
		Score: 818290,
		ClearStatus: false,
		TimePlayed: date,
		SongId: "b1do8OI6qDDlQO0PI16868ql6bdbI886",
		Mode: "SINGLE",
		Difficulty: "EXPERT",
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-18 18:49:09", timeLocation)
	expectedRecentScores = append(expectedRecentScores, ddr_models.Score{
		Score: 987390,
		ClearStatus: true,
		TimePlayed: date,
		SongId: "08PO96OlIoQqPdq91Q1Qqlo8lPidbPP8",
		Mode: "SINGLE",
		Difficulty: "EXPERT",
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-18 18:46:39", timeLocation)
	expectedRecentScores = append(expectedRecentScores, ddr_models.Score{
		Score: 989630,
		ClearStatus: true,
		TimePlayed: date,
		SongId: "1OlD9Iqb9Oqol09dDi6iiQ9Iod1oP0il",
		Mode: "SINGLE",
		Difficulty: "EXPERT",
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-18 18:43:30", timeLocation)
	expectedRecentScores = append(expectedRecentScores, ddr_models.Score{
		Score: 946550,
		ClearStatus: true,
		TimePlayed: date,
		SongId: "8QbqP80q9PI8bbi0qOoiibOQD08OPdli",
		Mode: "SINGLE",
		Difficulty: "DIFFICULT",
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-18 18:39:36", timeLocation)
	expectedRecentScores = append(expectedRecentScores, ddr_models.Score{
		Score: 835080,
		ClearStatus: true,
		TimePlayed: date,
		SongId: "bilO9D91P1lo81oIDP6qOoqdDdQdoDlP",
		Mode: "SINGLE",
		Difficulty: "EXPERT",
		PlayerCode: 12345678,
	})

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	recentScores, err := recentScoresFromDocument(document, 12345678)

	if err != nil {
		t.Errorf("failed to load chart stats from document: %s", err.Error())
	}

	for _, rs := range recentScores {
		match := false
		for _, ers := range expectedRecentScores {
			if rs.PlayerCode == ers.PlayerCode &&
				rs.Score == ers.Score &&
				rs.ClearStatus == ers.ClearStatus &&
				rs.TimePlayed.String() == ers.TimePlayed.String() &&
				rs.SongId == ers.SongId &&
				rs.Mode == ers.Mode &&
				rs.Difficulty == ers.Difficulty {
				match = true
				break
			}
		}
		if !match {
			t.Errorf("could not find a match for score: %+#v", rs)
		}
	}
}

func TestRecentScoresForClient(t *testing.T) {

}

func TestWorkoutDataFromDocument(t *testing.T) {
	// Setup test
	const testFile = "./test_data/player/workout.html"

	// Setup expected results
	timeFormat := "2006-01-02"
	timeLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		glog.Warningln("failed to load timezone location Asia/Tokyo")
		return
	}

	expectedWorkoutData := make([]ddr_models.WorkoutData, 0)

	date, _ := time.ParseInLocation(timeFormat, "2020-03-18", timeLocation)
	expectedWorkoutData = append(expectedWorkoutData, ddr_models.WorkoutData{
		Date:       date,
		PlayCount:  11,
		Kcal:       343.35,
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-16", timeLocation)
	expectedWorkoutData = append(expectedWorkoutData, ddr_models.WorkoutData{
		Date:       date,
		PlayCount:  17,
		Kcal:       613.958,
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-11", timeLocation)
	expectedWorkoutData = append(expectedWorkoutData, ddr_models.WorkoutData{
		Date:       date,
		PlayCount:  12,
		Kcal:       412.469,
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-06", timeLocation)
	expectedWorkoutData = append(expectedWorkoutData, ddr_models.WorkoutData{
		Date:       date,
		PlayCount:  7,
		Kcal:       207.611,
		PlayerCode: 12345678,
	})

	date, _ = time.ParseInLocation(timeFormat, "2020-03-05", timeLocation)
	expectedWorkoutData = append(expectedWorkoutData, ddr_models.WorkoutData{
		Date:       date,
		PlayCount:  17,
		Kcal:       538.68,
		PlayerCode: 12345678,
	})

	// Run Test
	document, err := documentFromFile(testFile)
	if err != nil {
		t.Fatalf("could not load %s: %s", testFile, err.Error())
	}

	workoutData, err := workoutDataFromDocument(document, 12345678)

	if err != nil {
		t.Errorf("failed to load chart stats from document: %s", err.Error())
	}

	for _, wd := range workoutData {
		match := false
		for _, ewd := range expectedWorkoutData {
			if wd.PlayerCode == ewd.PlayerCode &&
				wd.PlayCount == ewd.PlayCount &&
				wd.Kcal == ewd.Kcal &&
				wd.Date.String() == ewd.Date.String() {
				match = true
				break
			}
		}
		if !match {
			t.Errorf("could not find a match for workout data: %+#v", wd)
		}
	}
}

func TestWorkoutDataForClient(t *testing.T) {

}