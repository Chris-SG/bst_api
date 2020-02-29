package main

import (
	"encoding/json"
	"github.com/chris-sg/eagate/ddr"
	"github.com/chris-sg/eagate/user"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/ddr_db"
	"github.com/chris-sg/eagate_db/user_db"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

func CreateDdrRouter() *mux.Router {
	ddrRouter := mux.NewRouter().PathPrefix("/ddr").Subrouter()

	ddrRouter.HandleFunc("/songs", SongsGet).Methods(http.MethodGet)
	ddrRouter.Path("/songs").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsPatch)))).Methods(http.MethodPatch)

	return ddrRouter
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