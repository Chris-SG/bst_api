package drs

import (
	"github.com/chris-sg/bst_api/models/drs_models"
	"github.com/golang/glog"
	"strings"
	"time"
)

func Transform(dancerInfo drs_models.DancerInfo, musicData drs_models.MusicData, playHist drs_models.PlayHist) (pd drs_models.PlayerDetails, pps drs_models.PlayerProfileSnapshot, s []drs_models.Song, d []drs_models.Difficulty, pss []drs_models.PlayerSongStats, ps []drs_models.PlayerScore) {
	pd = drs_models.PlayerDetails{
		Code:       musicData.Data.PlayerData.UserId.Code,
		Name:       dancerInfo.Data.EaSite.Profile.Name,
		EaGateUser: nil,
	}
	
	pps = drs_models.PlayerProfileSnapshot{
		PlayCount:   dancerInfo.Data.EaSite.Statistics.PlayCount,
		PlaySeconds: dancerInfo.Data.EaSite.Statistics.PlaySecs,
		TotalStars:  dancerInfo.Data.EaSite.Coins.Total,
		UsedStars:   dancerInfo.Data.EaSite.Coins.Used,
		LastPlayed:  time.Unix(0, playHist.Data.PlayerData.MusicHistory.Music[0].LastPlayDate*1000000),
		PlayerCode:  musicData.Data.PlayerData.UserId.Code,
	}
	
	for songId, songDetails := range musicData.Data.PlayerData.MusicDb {
		song := drs_models.Song{
			SongId:         songId,
			SongName:       songDetails.Info.TitleName,
			ArtistName:     songDetails.Info.ArtistName,
			MaxBpm:         songDetails.Info.BpmMax,
			MinBpm:         songDetails.Info.BpmMin,
			LimitationType: songDetails.Info.LimitationType,
			Genre:          songDetails.Info.Genre,
			VideoFlags:     songDetails.Info.PlayVideoFlags,
			License:        songDetails.Info.License,
		}
		s = append(s, song)
		for diffType, rawDiff := range songDetails.Difficulties {
			if !strings.HasPrefix(diffType, "fumen_") {
				glog.Warningf("diff field for %s does not have prefix, instead %s\n", songId, diffType)
				continue
			}
			diffType = diffType[len("fumen_"):]
			mode := "Double"
			if strings.HasPrefix(diffType, "1") {
				mode = "Single"
			}
			difficulty := "Easy"
			if strings.HasSuffix(diffType, "a") {
				difficulty = "Normal"
			}

			diff := drs_models.Difficulty{
				Mode:        mode,
				Difficulty: difficulty,
				Level:      rawDiff.DiffNum,
				SongId:     songId,
			}
			d = append(d, diff)
		}
	}

	for _, chart := range musicData.Data.PlayerData.ScoreData.Music {
		mode := "Double"
		if strings.HasPrefix(chart.MusicType, "1") {
			mode = "Single"
		}
		difficulty := "Easy"
		if strings.HasSuffix(chart.MusicType, "a") {
			difficulty = "Normal"
		}
		
		stat := drs_models.PlayerSongStats{
			BestScore:         chart.Score,
			Combo:             chart.Combo,
			PlayCount:         chart.PlayCount,
			Param:             chart.Param,
			BestScoreDateTime: time.Unix(0, chart.BestScoreDate*1000000),
			LastPlayDateTime:  time.Unix(0, chart.LastPlayDate*1000000),
			P1Code:            chart.Player1.Code,
			P1Score:           chart.Player1.Score,
			P1Perfects:        chart.Player1.Perfect,
			P1Greats:          chart.Player1.Great,
			P1Goods:           chart.Player1.Good,
			P1Bads:            chart.Player1.Bad,
			P2Code:            nil,
			P2Score:           nil,
			P2Perfects:        nil,
			P2Greats:          nil,
			P2Goods:           nil,
			P2Bads:            nil,
			PlayerCode:        musicData.Data.PlayerData.UserId.Code,
			SongId:            chart.MusicId,
			Mode:              mode,
			Difficulty:        difficulty,
		}
		
		if chart.Player2 != nil {
			stat.P2Code = &chart.Player2.Code
			stat.P2Score = &chart.Player2.Score
			stat.P2Perfects = &chart.Player2.Perfect
			stat.P2Greats = &chart.Player2.Great
			stat.P2Goods = &chart.Player2.Good
			stat.P2Bads = &chart.Player2.Bad
		}
		
		pss = append(pss, stat)
	}
	
	for _, score := range playHist.Data.PlayerData.MusicHistory.Music {
		mode := "Double"
		if strings.HasPrefix(score.MusicType, "1") {
			mode = "Single"
		}
		difficulty := "Easy"
		if strings.HasSuffix(score.MusicType, "a") {
			difficulty = "Normal"
		}

		recentScore := drs_models.PlayerScore{
			Shop:       score.ShopName,
			Score:      score.Score,
			MaxCombo:   score.Combo,
			Param:      score.Param,
			PlayTime:   time.Unix(0, score.LastPlayDate*1000000),
			P1Code:     score.Player1.PlayerCode,
			P1Score:    score.Player1.MemberScore,
			P1Perfects: score.Player1.Perfect,
			P1Greats:   score.Player1.Great,
			P1Goods:    score.Player1.Good,
			P1Bads:     score.Player1.Bad,
			P2Code:     nil,
			P2Score:    nil,
			P2Perfects: nil,
			P2Greats:   nil,
			P2Goods:    nil,
			P2Bads:     nil,
			VideoUrl:   nil,
			PlayerCode: musicData.Data.PlayerData.UserId.Code,
			SongId:     score.MusicId,
			Mode:       mode,
			Difficulty: difficulty,
		}

		if score.Player2 != nil {
			recentScore.P2Code = &score.Player2.PlayerCode
			recentScore.P2Score = &score.Player2.MemberScore
			recentScore.P2Perfects = &score.Player2.Perfect
			recentScore.P2Greats = &score.Player2.Great
			recentScore.P2Goods = &score.Player2.Good
			recentScore.P2Bads = &score.Player2.Bad
		}

		if len(score.VideoUrl) > 0 {
			recentScore.VideoUrl = &score.VideoUrl
		}

		ps = append(ps, recentScore)
	}

	return
}
