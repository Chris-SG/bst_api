package drs_models

// dancer_info
type DancerInfo struct {
	Status int `json:"status"`
	Data dancerInfoData `json:"data"`
}

type dancerInfoData struct {
	Status int `json:"status"`
	EaSite dancerInfoEaSite `json:"easite_get_playerdata"`
}

type dancerInfoEaSite struct {
	Result string `json:"result"`
	Profile dancerInfoProfile `json:"profile"`
	Statistics dancerInfoPlayStatistics `json:"statics_play"`
	Coins dancerInfoDanceCoin `json:"normal_dance_coin"`
	Camp dancerInfoCamp `json:"camp"`
}

type dancerInfoProfile struct {
	Name string `json:"name"`
}

type dancerInfoPlayStatistics struct {
	PlayCount int `json:"play_cnt"`
	PlaySecs int `json:"play_sec"`
}

type dancerInfoDanceCoin struct {
	Total int `json:"total"`
	Used int `json:"used"`
	Limit int `json:"limit"`
}

type dancerInfoCamp struct {
	VoteRights1 int `json:"vote_rights_1"`
	VoteRights2 int `json:"vote_rights_2"`
}


// music_data
type MusicData struct {
	Status int `json:"status"`
	Data musicDataData `json:"data"`
}

type musicDataData struct {
	Status int `json:"status"`
	PlayerData musicDataEaSitePlayerData `json:"easite_get_playerdata"`
}

type musicDataEaSitePlayerData struct {
	Result int `json:"result"`
	UserId musicDataUserId `json:"userid"`
	UnlockedMusic musicDataUnlockMusic `json:"unlock_music"`
	ScoreData musicDataScoreData `json:"scoredata"`
	MusicDb map[string]songDbEntry `json:"mdb"`
}

type musicDataUserId struct {
	Code int `json:"code"`
}

type musicDataUnlockMusic struct {
	MusicIds []string `json:"music_id"`
}

type musicDataScoreData struct {
	Music []musicDataMusic `json:"music"`
}

type musicDataMusic struct {
	MusicId string `json:"music_id"`
	MusicType string `json:"music_type"`
	PlayCount int `json:"play_cnt"`
	Score int `json:"score"`
	Rank int `json:"rank"`
	Combo int `json:"combo"`
	Param int `json:"param"`
	BestScoreDate int64 `json:"bestscore_date"`
	LastPlayDate int64 `json:"lastplay_date"`
	ShopName string `json:"shopname"`
	Player1 musicDataPlayerScoreDetails `json:"p1"`
	Player2 *musicDataPlayerScoreDetails `json:"p2,omitempty"`
}

type musicDataPlayerScoreDetails struct {
	Code int `json:"member_code"`
	Score int `json:"member_score"`
	Perfect int `json:"perfect"`
	Great int `json:"great"`
	Good int `json:"good"`
	Bad int `json:"bad"`
}


// play_hist
type PlayHist struct {
	Status int `json:"status"`
	Data playHistData `json:"data"`
}

type playHistData struct {
	Status int `json:"status"`
	PlayerData playHistEaSitePlayerData `json:"easite_get_playerdata"`
}

type playHistEaSitePlayerData struct {
	Result int `json:"result"`
	UserId playHistUserId `json:"userid"`
	MusicHistory playHistMusicHist `json:"music_hist"`
	MusicDb map[string]songDbEntry `json:"mdb"`
}

type playHistUserId struct {
	Code int `json:"code"`
}

type playHistMusicHist struct {
	Music []playHistMusic `json:"music"`
}

type playHistMusic struct {
	StageNumber int `json:"stage_no"`
	MusicId string `json:"music_id"`
	MusicType string `json:"music_type"`
	PlayCount int `json:"play_cnt"`
	Score int `json:"score"`
	Rank int `json:"rank"`
	Combo int `json:"combo"`
	Param int `json:"param"`
	BestScoreDate int64 `json:"bestscore_date"`
	LastPlayDate int64 `json:"lastplay_date"`
	ShopName string `json:"shopname"`
	Player1 playHistPlayerScoreDetails `json:"p1"`
	Player2 *playHistPlayerScoreDetails `json:"p2,omitempty"`
	VideoUrl string `json:"video_url"`
}

type playHistPlayerScoreDetails struct {
	PlayerCode int `json:"member_code"`
	MemberScore int `json:"member_score"`
	Perfect int `json:"perfect"`
	Great int `json:"great"`
	Good int `json:"good"`
	Bad int `json:"bad"`
}


// song db entries
type songDbEntry struct {
	Info songInfo `json:"info"`
	Difficulties map[string]songDifficulty `json:"difficulty"`
}

type songInfo struct {
	MusicId string `json:"music_id"`
	TitleName string `json:"title_name"`
	TitleYomigana string `json:"title_yomigana"`
	ArtistName string `json:"artist_name"`
	ArtistYomigana string `json:"artist_yomigana"`
	BpmMax int `json:"bpm_max"`
	BpmMin int `json:"bpm_min"`
	LimitationType int `json:"limitation_type"`
	Genre int `json:"genre"`
	PlayVideoFlags int `json:"play_video_flags"`
	License string `json:"license"`
}

type songDifficulty struct {
	DiffNum int `json:"difnum"`
	Playable int `json:"playable"`
}