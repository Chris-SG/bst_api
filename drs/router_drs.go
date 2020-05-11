package drs

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/utilities"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

// CreateDrsRouter will create a mux router to be attached to
// the main router, prefixed with '/ddr'.
func CreateDrsRouter() *mux.Router {
	drsRouter := mux.NewRouter().PathPrefix("/drs").Subrouter()

	drsRouter.Path("/profile").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfilePatch)))).Methods(http.MethodPatch)

	drsRouter.Path("/details").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(DetailsGet)))).Methods(http.MethodGet)

	drsRouter.Path("/songs/stats").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongStatsGet)))).Methods(http.MethodGet)

	drsRouter.Path("/tabledata").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(TableDataGet)))).Methods(http.MethodGet)

	return drsRouter
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func ProfilePatch(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	for _, user := range users {
		client, errMsg, err := common.CreateClientForUser(user)
		if err != nil {
			status := utilities.WriteStatus("bad", errMsg)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(bytes)
			return
		}

		err = refreshDrsUser(client)
		if err != nil {
			status := utilities.WriteStatus("bad", err.Error())
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(bytes)
			return
		}
	}

	status := utilities.WriteStatus("ok", "profile refreshed")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func DetailsGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	if len(users) == 0 {
		status := utilities.WriteStatus("bad", "drs_nouser")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	details, err := retrieveDrsPlayerDetails(users[0].Name)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(details)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func SongStatsGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	if len(users) == 0 {
		status := utilities.WriteStatus("bad", "drs_nouser")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	details, err := retrieveDrsPlayerDetails(users[0].Name)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	stats, err := retrieveDrsSongStats(details.Code)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(stats)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func TableDataGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	if len(users) == 0 {
		status := utilities.WriteStatus("bad", "drs_nouser")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	details, err := retrieveDrsPlayerDetails(users[0].Name)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	tableData, errs := eagate_db.GetDrsDb().RetrieveDataForTable(details.Code)
	if utilities.PrintErrors("failed to retrieve user:", errs) {
		status := utilities.WriteStatus("bad", "drs_rettbl_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(tableData))
	return
}