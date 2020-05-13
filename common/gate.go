package common

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/bst_models"
	"github.com/chris-sg/bst_api/utilities"
	bstServerModels "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
	"net/http"
	"time"
)

var (
	nextUpdate time.Time
	cachedGate bool
	cachedDb bool
)

// Status will return any status details for the API. This
// currently caches eagate and db connections, however these
// are still regenerated every page load.
// TODO: use proper caching with timed expiry.
func Status(rw http.ResponseWriter, r *http.Request) {
	if nextUpdate.Unix() < time.Now().Unix() {
		nextUpdate = time.Now().Add(time.Minute * 2)
		updateCachedDb()
		updateCachedGate()
	}

	status := bstServerModels.ApiStatus{
		Api: "ok",
	}
	if cachedGate {
		status.EaGate = "ok"
	} else {
		status.EaGate = "bad"
	}
	if cachedDb {
		status.Db = "ok"
	} else {
		status.Db = "bad"
	}

	statusBytes, _ := json.Marshal(status)

	rw.WriteHeader(http.StatusOK)
	rw.Write(statusBytes)
}

// updateCachedDb will retrieve the current database status.
// This allows us to confirm whether the connection has broken
// or not.
func updateCachedDb() {
	db, err := db.GetDb()
	if err != nil || db.DB().Ping() != nil {
		cachedDb = false
	} else {
		cachedDb = true
	}
}

// updateCachedGate will get the current state of eagate. This
// will return false if either a connection cannot be formed
// with eagate or maintenance mode is active.
func updateCachedGate() {
	client := util.GenerateClient()
	cachedGate = !util.IsMaintenanceMode(client)
}

func Cache(rw http.ResponseWriter, r *http.Request) {
	data := bstServerModels.UserCache{}

	query := r.URL.Query()
	user := query.Get("user")
	glog.Infof("get cache data for %s", user)

	apiDb := db.GetApiDb()

	profile, errs := apiDb.RetrieveProfile(user)
	if utilities.PrintErrors("failed to retrieve profile:", errs) {
		status := utilities.WriteStatus("bad", "cache_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}
	if len(profile.User) == 0 {
		glog.Infof("user %s not found, adding to db", user)
		profile = bst_models.BstProfile{ User: user, Public: false }
		errs = apiDb.SetProfile(profile)
		if utilities.PrintErrors("failed to set profile:", errs) {
			status := utilities.WriteStatus("bad", "cache_err")
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write(bytes)
			return
		}
	}

	data.Public = profile.Public
	data.Nickname = profile.Nickname
	data.Id = profile.UserId

	bytes, _ := json.Marshal(data)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

func PutBstUser(rw http.ResponseWriter, r *http.Request) {
	type UpdateableData struct {
		Nickname string `json:"nickname"`
		Public bool `json:"public"`
	}

	tokenMap := utilities.ProfileFromToken(r)

	user, ok := tokenMap["sub"].(string)
	if !ok {
		status := utilities.WriteStatus("bad", "jwt_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	data := UpdateableData{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		glog.Warningf("unable to decode body: %s", err.Error())
		status := utilities.WriteStatus("bad", "invalid_data")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	apiDb := db.GetApiDb()
	profile, errs := apiDb.RetrieveProfile(user)
	if utilities.PrintErrors("failed to retrieve profile:", errs) {
		status := utilities.WriteStatus("bad", "bst_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}
	if len(profile.User) == 0 {
		profile.User = user
	}
	if len(data.Nickname) > 0 {
		profile.Nickname = data.Nickname
	}
	profile.Public = data.Public

	errs = apiDb.SetProfile(profile)
	if utilities.PrintErrors("failed to set profile:", errs) {
		status := utilities.WriteStatus("bad", "bst_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(profile)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}