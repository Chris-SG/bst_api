package drs_models

import "time"

type PlayerDetails struct {
	Code int    `gorm:"column:code;primary_key" json:"code"`
	Name string `gorm:"column:name" json:"name"`

	EaGateUser *string `gorm:"column:eagate_user" json:"eagateuser"`
}

func (PlayerDetails) TableName() string {
	return "drsPlayerDetails"
}

type PlayerProfileSnapshot struct {
	PlayCount   int `gorm:"column:play_count;primary_key" json:"playcount"`
	PlaySeconds int `gorm:"column:play_seconds" json:"playseconds"`
	TotalStars  int `gorm:"column:total_stars" json:"totalstars"`
	UsedStars   int `gorm:"column:used_stars" json:"usedstars"`
	LastPlayed  time.Time `gorm:"column:last_played" json:"timeplayed"`

	PlayerCode int `gorm:"column:player_code;primary_key" json:"code"`
}

func (PlayerProfileSnapshot) TableName() string {
	return "drsPlayerProfileSnapshots"
}

type Song struct {
	SongId         string `gorm:"column:song_id;primary_key" json:"id"`
	SongName       string `gorm:"column:name" json:"title"`
	ArtistName     string `gorm:"column:artist" json:"artist"`
	MaxBpm         int    `gorm:"column:max_bpm" json:"maxbpm"`
	MinBpm         int    `gorm:"column:min_bpm" json:"minbpm"`
	LimitationType int    `gorm:"column:limitation_type" json:"limitation"`
	Genre          int    `gorm:"column:genre" json:"genre"`
	VideoFlags     int    `gorm:"column:video_flags" json:"videoflags"`
	License        string `gorm:"column:license" json:"license"`
}

func (Song) TableName() string {
	return "drsSongs"
}

type Difficulty struct {
	Mode       string `gorm:"column:mode;primary_key" json:"mode"`
	Difficulty string `gorm:"column:difficulty;primary_key" json:"difficulty"`
	Level      int    `gorm:"column:level" json:"level"`

	SongId string `gorm:"column:song_id;primary_key" json:"id"`
}

func (Difficulty) TableName() string {
	return "drsDifficulties"
}

type PlayerSongStats struct {
	BestScore         int       `gorm:"column:best_score" json:"bestscore"`
	Combo             int       `gorm:"column:combo" json:"combo"`
	PlayCount         int       `gorm:"column:play_count" json:"playcount"`
	Param             int       `gorm:"column:param" json:"param"`
	BestScoreDateTime time.Time `gorm:"column:best_score_time" json:"bestscoretime"`
	LastPlayDateTime  time.Time `gorm:"column:last_play_time" json:"lastplaytime"`

	P1Code     int `gorm:"column:p1_code" json:"p1code"`
	P1Score    int `gorm:"column:p1_score" json:"p1combo"`
	P1Perfects int `gorm:"column:p1_perfects" json:"p1perfects"`
	P1Greats   int `gorm:"column:p1_greats" json:"p1greats"`
	P1Goods    int `gorm:"column:p1_goods" json:"p1goods"`
	P1Bads     int `gorm:"column:p1_bads" json:"p1bads"`

	P2Code     *int `gorm:"column:p2_code" json:"p2code;omitempty"`
	P2Score    *int `gorm:"column:p2_score" json:"p2score;omitempty"`
	P2Perfects *int `gorm:"column:p2_perfects" json:"p2perfects;omitempty"`
	P2Greats   *int `gorm:"column:p2_greats" json:"p2greats;omitempty"`
	P2Goods    *int `gorm:"column:p2_goods" json:"p2goods;omitempty"`
	P2Bads     *int `gorm:"column:p2_bads" json:"p2bads;omitempty"`

	PlayerCode int    `gorm:"column:player_code;primary_key" json:"code"`
	SongId     string `gorm:"column:song_id;primary_key" json:"id"`
	Mode       string `gorm:"column:mode;primary_key" json:"mode"`
	Difficulty string `gorm:"column:difficulty;primary_key" json:"difficulty"`
}

func (PlayerSongStats) TableName() string {
	return "drsPlayerSongStats"
}

func (s1 PlayerSongStats) Equals(s2 PlayerSongStats) bool {
	return s1.SongId == s2.SongId &&
		s1.Mode == s2.Mode &&
		s1.Difficulty == s2.Difficulty &&
		s1.PlayerCode == s2.PlayerCode &&
		s1.BestScoreDateTime.String() == s2.BestScoreDateTime.String() &&
		s1.LastPlayDateTime.String() == s2.LastPlayDateTime.String() &&
		s1.BestScore == s2.BestScore &&
		s1.Combo == s2.Combo &&
		s1.PlayCount == s2.PlayCount
}

type PlayerScore struct {
	Shop     string    `gorm:"column:shop" json:"shop"`
	Score    int       `gorm:"column:score" json:"score"`
	MaxCombo int       `gorm:"column:max_combo" json:"maxcombo"`
	Param    int       `gorm:"column:param" json:"param"`
	PlayTime time.Time `gorm:"column:play_time" json:"playtime"`

	P1Code     int `gorm:"column:p1_code" json:"p1code"`
	P1Score    int `gorm:"column:p1_score" json:"p1score"`
	P1Perfects int `gorm:"column:p1_perfects" json:"p1perfects"`
	P1Greats   int `gorm:"column:p1_greats" json:"p1greats"`
	P1Goods    int `gorm:"column:p1_goods" json:"p1goods"`
	P1Bads     int `gorm:"column:p1_bads" json:"p1bads"`

	P2Code     *int `gorm:"column:p2_code" json:"p2code;omitempty"`
	P2Score    *int `gorm:"column:p2_score" json:"p2score;omitempty"`
	P2Perfects *int `gorm:"column:p2_perfects" json:"p2perfects;omitempty"`
	P2Greats   *int `gorm:"column:p2_greats" json:"p2greats;omitempty"`
	P2Goods    *int `gorm:"column:p2_goods" json:"p2goods;omitempty"`
	P2Bads     *int `gorm:"column:p2_bads" json:"p2bads;omitempty"`

	VideoUrl *string `gorm:"column:video_url" json:"videourl;omitempty"`

	PlayerCode int    `gorm:"column:player_code;primary_key" json:"code"`
	SongId     string `gorm:"column:song_id;primary_key" json:"id"`
	Mode       string `gorm:"column:mode;primary_key" json:"mode"`
	Difficulty string `gorm:"column:difficulty;primary_key" json:"difficulty"`
}

func (PlayerScore) TableName() string {
	return "drsPlayerScores"
}
