package main

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_server_models/bst_web_models"
	"github.com/chris-sg/eagate/ddr"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/ddr_db"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
)

// CreateDdrRouter will create a mux router to be attached to
// the main router, prefixed with '/ddr'.
func CreateDdrRouter() *mux.Router {
	ddrRouter := mux.NewRouter().PathPrefix("/ddr").Subrouter()

	ddrRouter.Path("/profile/update").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(ProfileUpdatePatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/profile/refresh").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(ProfileRefreshPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs", SongsGet).Methods(http.MethodGet)
	ddrRouter.Path("/songs").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs/jackets", SongsJacketGet).Methods(http.MethodGet)

	ddrRouter.HandleFunc("/songs/{id}", SongsIdGet).Methods(http.MethodGet)

	ddrRouter.Path("/songs/scores").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsScoresGet)))).Methods(http.MethodGet)

	return ddrRouter
}

// ProfileRefreshPatch will perform a full refresh of all song
// difficulties in the database for the user. This is an expensive
// operation and should be used sparingly.
func ProfileRefreshPatch(rw http.ResponseWriter, r *http.Request) {
	users, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	db, _ := eagate_db.GetDb()
	difficulties := ddr_db.RetrieveValidSongDifficulties(db)

	for _, user := range users {
		client, err := createClientForUser(user)
		if err != nil {
			status := WriteStatus("bad", err.Error())
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(bytes)
			return
		}

		err = updateSongStatistics(client, difficulties)
		if err != nil {
			status := WriteStatus("bad", err.Error())
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(bytes)
			return
		}
	}

	status := WriteStatus("ok", "profile refreshed")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// ProfileUpdatePatch will check the past 50 plays for the user.
// These scores will be added to the database, and then the
// difficulty details will be updated for the user. This should
// be used in favour of ProfileRefreshPatch where possible.
func ProfileUpdatePatch(rw http.ResponseWriter, r *http.Request) {
	users, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	for _, user := range users {
		client, err := createClientForUser(user)
		if err != nil {
			status := WriteStatus("bad", err.Error())
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(bytes)
			return
		}

		err = updatePlayerProfile(user, client)
		if err != nil {
			status := WriteStatus("bad", err.Error())
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(bytes)
			return
		}
	}

	status := WriteStatus("ok", "profile updated")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongsGet will retrieve a list of all songs from the database.
// Data returned will not include the jacket image, which should
// be retrieved with the `/ddr/songs/images` endpoint.
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

// SongsPatch will attempt to find which songs on eagate are not
// yet in the database, and proceed to add songs as their difficulties
// to the database.
// TODO: if this fails after adding the songs to the database, new
// difficulties will be missing with no current recovery method.
func SongsPatch(rw http.ResponseWriter, r *http.Request) {
	users, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	client, err := createClientForUser(users[0])
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	newSongs, err := checkForNewSongs(client)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	db, _ := eagate_db.GetDb()
	songData, err := ddr.SongData(client, newSongs)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	ddr_db.AddSongs(db, songData)

	songDifficulties, err := ddr.SongDifficulties(client, newSongs)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	ddr_db.AddSongDifficulties(db, songDifficulties)

	status := WriteStatus("ok", fmt.Sprintf("added %d new songs (%d new difficulties)", len(newSongs), len(songDifficulties)))
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongJacketsGet will retrieve the jackets for song ids provided in the
// request body
func SongsJacketGet(rw http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	ids := bst_web_models.DdrSongIds{}
	err = json.Unmarshal(body, &ids)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	songs := ddr_db.RetrieveSongsWithCovers(db, ids.Ids)

	jackets := make([]bst_web_models.DdrSongIdWithJacket, len(songs))
	for i, _ := range songs {
		jackets[i].Id = songs[i].Id
		jackets[i].Jacket = songs[i].Image
	}

	bytes, _ := json.Marshal(jackets)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func SongsIdGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val := vars["id"]

	db, err := eagate_db.GetDb()
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	songs := ddr_db.RetrieveSongsWithCovers(db, []string{val})
	if len(songs) == 0 {
		status := WriteStatus("bad", fmt.Sprintf("unable to find song id %s in database", val))
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	difficulties := ddr_db.RetrieveSongDifficultiesById(db, []string{val})
	if len(difficulties) == 0 {
		status := WriteStatus("bad", fmt.Sprintf("unable to find difficulties for song id %s in database", val))
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	songs[0].Difficulties = difficulties

	bytes, _ := json.Marshal(difficulties)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func SongsScoresGet(rw http.ResponseWriter, r *http.Request) {
	users, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	db, _ := eagate_db.GetDb()

	ddrProfile, err := ddr_db.RetrieveDdrPlayerDetailsByEaGateUser(db, users[0].Name)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	scores := ddr_db.RetrieveSongStatistics(db, ddrProfile.Code)

	bytes, _ := json.Marshal(scores)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}