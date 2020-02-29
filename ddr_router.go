package main

import (
	"encoding/json"
	"github.com/chris-sg/eagate/ddr"
	"github.com/chris-sg/eagate/user"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/ddr_db"
	"github.com/chris-sg/eagate_db/user_db"
	"github.com/chris-sg/eagate_models/ddr_models"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

func CreateDdrRouter() *mux.Router {
	ddrRouter := mux.NewRouter().PathPrefix("/ddr").Subrouter()

	ddrRouter.Path("/profile/update").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(ProfileUpdatePatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/profile/refresh").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(ProfileRefreshPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs", SongsGet).Methods(http.MethodGet)
	ddrRouter.Path("/songs").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsPatch)))).Methods(http.MethodPatch)

	return ddrRouter
}

func ProfileRefreshPatch(rw http.ResponseWriter, r *http.Request) {
	profile := profileFromToken(r)

	val, ok := profile["name"].(string)
	if !ok {
		rw.Write([]byte(`name not found`))
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		panic(err)
	}

	users := user_db.RetrieveUserByWebId(db, val)
	if len(users) == 0 {
		rw.WriteHeader(http.StatusPreconditionFailed)
		rw.Write([]byte(`{"error":"user not logged into eagate"}`))
		return
	}

	client := util.GenerateClient()
	cookie := user_db.RetrieveUserCookieById(db, users[0].Name)
	err = user.CheckCookieEaGateAccess(client, cookie)
	if err != nil {
		rw.WriteHeader(http.StatusPreconditionFailed)
		rw.Write([]byte(`{"error":"cookie error"}`))
		return
	}

	user.AddCookiesToJar(client.Client.Jar, []*http.Cookie{cookie})
	pi, _, _ := ddr.PlayerInformation(client)
	ids, _ := ddr.SongIds(client)
	difficulties, _ := ddr.SongDifficulties(client, ids)
	stats, _ := ddr.SongStatistics(client, difficulties, pi.Code)

	ddr_db.AddSongStatistics(db, stats, pi.Code)
}

func ProfileUpdatePatch(rw http.ResponseWriter, r *http.Request) {
	profile := profileFromToken(r)

	val, ok := profile["name"].(string)
	if !ok {
		rw.Write([]byte(`name not found`))
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		panic(err)
	}

	users := user_db.RetrieveUserByWebId(db, val)
	if len(users) == 0 {
		rw.WriteHeader(http.StatusPreconditionFailed)
		rw.Write([]byte(`{"error":"user not logged into eagate"}`))
		return
	}

	client := util.GenerateClient()
	cookie := user_db.RetrieveUserCookieById(db, users[0].Name)
	err = user.CheckCookieEaGateAccess(client, cookie)
	if err != nil {
		rw.WriteHeader(http.StatusPreconditionFailed)
		rw.Write([]byte(`{"error":"cookie error"}`))
		return
	}

	user.AddCookiesToJar(client.Client.Jar, []*http.Cookie{cookie})
	newPi, playcount, _ := ddr.PlayerInformation(client)
	newPi.EaGateUser = &users[0].Name
	dbPi, _ := ddr_db.RetrieveDdrPlayerDetailsByEaGateUser(db, users[0].Name)
	if dbPi != nil {
		dbPlaycount := ddr_db.RetrieveLatestPlaycountDetails(db, dbPi.Code)
		if dbPlaycount != nil {
			if playcount.Playcount == dbPlaycount.Playcount {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(`{"status":"not changed"`))
				return
			}
		}
	}

	recentScores, _ := ddr.RecentScores(client, newPi.Code)

	ddr_db.AddPlayerDetails(db, *newPi)
	ddr_db.AddPlaycountDetails(db, *playcount)
	if recentScores != nil {
		ddr_db.AddScores(db, *recentScores)
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

	statistics, _ := ddr.SongStatistics(client, songsToUpdate, newPi.Code)
	ddr_db.AddSongStatistics(db, statistics, newPi.Code)
}

func SongsGet(rw http.ResponseWriter, r *http.Request) {
	db, _ := eagate_db.GetDb()

	songIds := ddr_db.RetrieveSongIds(db)
	songs := ddr_db.RetrieveSongsById(db, songIds)


	bytes, err := json.Marshal(songs)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func SongsPatch(rw http.ResponseWriter, r *http.Request) {
	profile := profileFromToken(r)

	val, ok := profile["name"].(string)
	if !ok {
		rw.Write([]byte(`name not found`))
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		panic(err)
	}

	users := user_db.RetrieveUserByWebId(db, val)
	if len(users) == 0 {
		rw.WriteHeader(http.StatusPreconditionFailed)
		rw.Write([]byte(`{"error":"user not logged into eagate"}`))
		return
	}

	client := util.GenerateClient()
	cookie := user_db.RetrieveUserCookieById(db, users[0].Name)
	err = user.CheckCookieEaGateAccess(client, cookie)
	if err != nil {
		rw.WriteHeader(http.StatusPreconditionFailed)
		rw.Write([]byte(`{"error":"cookie error"}`))
		return
	}

	user.AddCookiesToJar(client.Client.Jar, []*http.Cookie{cookie})
	ids, err := ddr.SongIds(client)
	songData, err := ddr.SongData(client, ids)
	ddr_db.AddSongs(db, songData)
	songDifficulties, err := ddr.SongDifficulties(client, ids)
	ddr_db.AddSongDifficulties(db, songDifficulties)

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{"status":"ok"}`))
}