package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

// CreateDrsRouter will create a mux router to be attached to
// the main router, prefixed with '/ddr'.
func CreateDrsRouter() *mux.Router {
	drsRouter := mux.NewRouter().PathPrefix("/drs").Subrouter()

	drsRouter.Path("/profile").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(DrsUpdateUser)))).Methods(http.MethodPatch)

	return drsRouter
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func DrsUpdateUser(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	for _, user := range users {
		client, errMsg, err := createClientForUser(user)
		if err != nil {
			status := WriteStatus("bad", errMsg)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(bytes)
			return
		}

		err = refreshDrsUser(client)
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
