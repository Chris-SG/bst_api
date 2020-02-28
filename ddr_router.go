package main

import (
	"encoding/json"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/ddr_db"
	"github.com/gorilla/mux"
	"net/http"
)

func CreateDdrRouter() *mux.Router {
	ddrRouter := mux.NewRouter().PathPrefix("/ddr").Subrouter()

	ddrRouter.HandleFunc("/songs", SongsGet).Methods(http.MethodGet)

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