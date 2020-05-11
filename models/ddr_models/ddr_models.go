package ddr_models

import (
	"time"
)

type PlayerDetails struct {
	Code        int    `tag:"DDR-CODE" gorm:"column:code;primary_key"`
	Name        string `tag:"ダンサーネーム" gorm:"column:name"`
	Prefecture  string `tag:"所属都道府県" gorm:"column:location"`
	SingleRank  string `tag:"段位(SINGLE)" gorm:"column:single_rank"`
	DoubleRank  string `tag:"段位(DOUBLE)" gorm:"column:double_rank"`
	Affiliation string `tag:"所属クラス" gorm:"column:affiliation"`

	EaGateUser *string          `gorm:"column:eagate_user"`
}

func (PlayerDetails) TableName() string {
	return "ddrPlayerDetails"
}

type Playcount struct {
	Playcount          int       `tag:"総プレー回数" gorm:"column:playcount"`
	LastPlayDate       time.Time `tag:"最終プレー日時" gorm:"column:last_play_date;primary_key"`
	SinglePlaycount    int       `tag:"プレー回数_single" gorm:"column:single_playcount"`
	SingleLastPlayDate time.Time `tag:"最終プレー日時_single" gorm:"column:last_single_play_date"`
	DoublePlaycount    int       `tag:"プレー回数_double" gorm:"column:double_playcount"`
	DoubleLastPlayDate time.Time `tag:"最終プレー日時_double" gorm:"column:last_double_play_date"`

	PlayerCode int           `gorm:"column:player_code;primary_key"`
}

func (Playcount) TableName() string {
	return "ddrPlaycount"
}

func (p1 Playcount) Equals(p2 Playcount) bool {
	return p1.PlayerCode == p2.PlayerCode &&
		p1.Playcount == p2.Playcount &&
		p1.LastPlayDate.String() == p2.LastPlayDate.String()
}

type Song struct {
	Id     string `gorm:"column:id;primary_key"`
	Name   string `gorm:"column:name"`
	Artist string `gorm:"column:artist"`
	Image  string `gorm:"column:image"`
}

func (Song) TableName() string {
	return "ddrSongs"
}

type SongDifficulty struct {
	SongId          string `gorm:"column:song_id;primary_key"`
	Mode            string `gorm:"column:mode;primary_key"`
	Difficulty      string `gorm:"column:difficulty;primary_key"`
	DifficultyValue int16  `gorm:"column:difficulty_value"`
}

func (SongDifficulty) TableName() string {
	return "ddrSongDifficulties"
}

type Mode int

const (
	Single Mode = iota
	Double
)

func (mode Mode) String() string {
	modeLabels := [...]string{
		"SINGLE",
		"DOUBLE",
	}

	return modeLabels[mode%2]
}

func StringToMode(mode string) Mode {
	modeLabels := map[string]Mode{
		"SINGLE": Single,
		"DOUBLE": Double,
	}
	return modeLabels[mode]
}

type Difficulty int

const (
	Beginner Difficulty = iota
	Basic
	Difficult
	Expert
	Challenge
)

func (difficulty Difficulty) String() string {
	difficultyLabels := [...]string{
		"BEGINNER",
		"BASIC",
		"DIFFICULT",
		"EXPERT",
		"CHALLENGE",
	}

	return difficultyLabels[difficulty%5]
}

func StringToDifficulty(difficulty string) Difficulty {
	difficultyLabels := map[string]Difficulty{
		"BEGINNER": Beginner,
		"BASIC": Basic,
		"DIFFICULT": Difficult,
		"EXPERT": Expert,
		"CHALLENGE": Challenge,
	}
	return difficultyLabels[difficulty]
}

type SongStatistics struct {
	BestScore  int       `tag:"ハイスコア" gorm:"column:score_record"`
	Lamp       string    `tag:"フルコンボ種別" gorm:"column:clear_lamp"`
	Rank       string    `tag:"ハイスコア時のダンスレベル" gorm"column:rank"`
	PlayCount  int       `tag:"プレー回数" gorm:"column:playcount"`
	ClearCount int       `tag:"クリア回数" gorm:"column:clearcount"`
	MaxCombo   int       `tag:"最大コンボ数" gorm:"column:maxcombo"`
	LastPlayed time.Time `tag:"最終プレー時間" gorm:"column:lastplayed"`

	SongId     string `gorm:"column:song_id;primary_key"`
	Mode       string `gorm:"column:mode;primary_key"`
	Difficulty string `gorm:"column:difficulty;primary_key"`

	PlayerCode int `gorm:"column:player_code;primary_key"`
}

func (SongStatistics) TableName() string {
	return "ddrSongStatistics"
}

func (s1 SongStatistics) Equals(s2 SongStatistics) bool {
	return s1.SongId == s2.SongId &&
			s1.Mode == s2.Mode &&
			s1.Difficulty == s2.Difficulty &&
			s1.PlayerCode == s2.PlayerCode &&
			s1.BestScore == s2.BestScore &&
			s1.Lamp == s2.Lamp &&
			s1.Rank == s2.Rank &&
			s1.PlayCount == s2.PlayCount &&
			s1.ClearCount == s2.ClearCount &&
			s1.MaxCombo == s2.MaxCombo &&
			s1.LastPlayed.String() == s2.LastPlayed.String()
}

type Score struct {
	Score       int       `gorm:"column:score"`
	ClearStatus bool      `gorm:"column:cleared"`
	TimePlayed  time.Time `gorm:"column:time_played;primary_key"`

	SongId     string `gorm:"column:song_id;primary_key"`
	Mode       string `gorm:"column:mode;primary_key"`
	Difficulty string `gorm:"column:difficulty;primary_key"`

	PlayerCode int `gorm:"column:player_code;primary_key"`
}

func (Score) TableName() string {
	return "ddrScores"
}

func (s1 Score) Equals(s2 Score) bool {
	return s1.SongId == s2.SongId &&
		s1.Mode == s2.Mode &&
		s1.Difficulty == s2.Difficulty &&
		s1.PlayerCode == s2.PlayerCode &&
		s1.Score == s2.Score &&
		s1.ClearStatus == s2.ClearStatus &&
		s1.TimePlayed.String() == s2.TimePlayed.String()
}

type WorkoutData struct {
	Date time.Time `gorm:"column:date;primary_key"`
	PlayCount int `gorm:"column:playcount"`
	Kcal float32 `gorm:"column:kcal"`

	PlayerCode int `gorm:"column:player_code;primary_key"`
}

func (WorkoutData) TableName() string {
	return "ddrWorkoutData"
}

func (w1 WorkoutData) Equals(w2 WorkoutData) bool {
	return w1.PlayerCode == w2.PlayerCode &&
			w1.Kcal == w2.Kcal &&
			w1.PlayCount == w2.PlayCount &&
			w1.Date.String() == w2.Date.String()
}