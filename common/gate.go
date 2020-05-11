package common

import (
	"encoding/json"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/util"
	bst_models "github.com/chris-sg/bst_server_models"
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

	status := bst_models.ApiStatus{
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
	db, err := eagate_db.GetDb()
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
	type CacheableData struct {
		Id int `json:"id"`
		Nickname string `json:"nickname"`
		Public bool `json:"public"`
	}

//	query := r.URL.Query()
//	user := query.Get("user")
}