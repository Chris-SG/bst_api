package drs

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/user"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	errCount := 0
	for _, username := range usernames {
		if !func() bool {
			userModel, exists, errs := db.GetUserDb().RetrieveUserByUserId(username)
			if !exists {
				glog.Warningf("user %s does not exist in db", username)
				return false
			}
			if utilities.PrintErrors("failed to retrieve user from db: ", errs) {
				return false
			}
			client, err := user.CreateClientForUser(userModel)
			defer client.UpdateCookie()
			if !err.Equals(bst_models.ErrorOK) {
				glog.Errorf("failed to create client: %s", err.Message)
				utilities.RespondWithError(rw, err)
				return false
			}

			err = refreshDrsUser(client)
			if !err.Equals(bst_models.ErrorOK) {
				glog.Errorf("failed to refresh user: %s", err.Message)
				utilities.RespondWithError(rw, err)
				return false
			}
			return true
		}() {
			errCount++
		}
	}
	if errCount > 0 {
		utilities.RespondWithError(rw, bst_models.ErrorDrsPlayerInfo)
	}

	utilities.RespondWithError(rw, bst_models.ErrorOK)
	return
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func DetailsGet(rw http.ResponseWriter, r *http.Request) {
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	if len(usernames) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	details, err := retrieveDrsPlayerDetails(usernames[0])
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}
	if len(usernames) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	details, err := retrieveDrsPlayerDetails(usernames[0])
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	if len(usernames) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	details, err := retrieveDrsPlayerDetails(usernames[0])
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