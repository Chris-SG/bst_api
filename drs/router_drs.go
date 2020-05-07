package drs

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/common"
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
		negroni.Wrap(http.HandlerFunc(DrsUpdateUser)))).Methods(http.MethodPatch)

	return drsRouter
}

// DrsUpdateUser will load all data provided by the Dance
// Rush API.
func DrsUpdateUser(rw http.ResponseWriter, r *http.Request) {
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
