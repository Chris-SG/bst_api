package drs

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/user"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
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
	users, err := common.TryGetEagateUsers(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	for _, u := range users {
		if !func() bool {
			client, err := user.CreateClientForUser(u)
			defer client.UpdateCookie()
			if !err.Equals(bst_models.ErrorOK) {
				utilities.RespondWithError(rw, err)
				return false
			}

			err = refreshDrsUser(client)
			if !err.Equals(bst_models.ErrorOK) {
				utilities.RespondWithError(rw, err)
				return false
			}
			return true
		}() {
			return
		}
	}

	utilities.RespondWithError(rw, bst_models.ErrorOK)
	return
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func DetailsGet(rw http.ResponseWriter, r *http.Request) {
	users, err := common.TryGetEagateUsers(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	if len(users) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	details, err := retrieveDrsPlayerDetails(users[0].Name)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
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
	users, err := common.TryGetEagateUsers(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}
	if len(users) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	details, err := retrieveDrsPlayerDetails(users[0].Name)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	stats, err := retrieveDrsSongStats(details.Code)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
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
	users, err := common.TryGetEagateUsers(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	if len(users) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	details, err := retrieveDrsPlayerDetails(users[0].Name)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	tableData, errs := db.GetDrsDb().RetrieveDataForTable(details.Code)
	if utilities.PrintErrors("failed to retrieve user:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDrsPlayerInfoDbRead)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(tableData))
	return
}