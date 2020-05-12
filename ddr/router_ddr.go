package ddr

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/ddr"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/chris-sg/bst_api/utilities"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
	"time"
)

// CreateDdrRouter will create a mux router to be attached to
// the main router, prefixed with '/ddr'.
func CreateDdrRouter() *mux.Router {
	ddrRouter := mux.NewRouter().PathPrefix("/ddr").Subrouter()

	ddrRouter.Path("/profile").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfileGet)))).Methods(http.MethodGet)

	ddrRouter.Path("/profile/workoutdata").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfileWorkoutDataGet)))).Methods(http.MethodGet)

	ddrRouter.Path("/profile/update").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfileUpdatePatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/profile/refresh").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfileRefreshPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs", SongsGet).Methods(http.MethodGet)

	ddrRouter.Path("/songs").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongsPatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/songs/reload").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongsReloadPatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/songs/scores").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongsScoresGet)))).Methods(http.MethodGet)

	ddrRouter.Path("/song/scores").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongScoresGet)))).Methods(http.MethodGet)
	ddrRouter.Path("/song/jacket").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongJacketGet)))).Methods(http.MethodGet)

	ddrRouter.Path("/songs/scores/extended").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongsScoresExtendedGet)))).Methods(http.MethodGet)

	return ddrRouter
}

// ProfileRefreshPatch will perform a full refresh of all song
// difficulties in the database for the user. This is an expensive
// operation and should be used sparingly.
func ProfileRefreshPatch(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	for _, user := range users {
		client, errMsg, err := common.CreateClientForUser(user)
		if err != nil {
			status := utilities.WriteStatus("bad", errMsg)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			_, _ = rw.Write(bytes)
			return
		}

		errMsg, err = refreshDdrUser(client)
		if err != nil {
			status := utilities.WriteStatus("bad", errMsg)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write(bytes)
			return
		}
	}

	status := utilities.WriteStatus("ok", "profile refreshed")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

// ProfileGet will retrieve formatted profile details for
// the current user.
func ProfileGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	playerDetails, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if utilities.PrintErrors("failed to retrieve player details by eagate user:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retpi_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	today := time.Now()
	tz, _ := time.LoadLocation("UTC")

	endDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, tz)
	startDate := endDate.AddDate(0, -1, 0)
	wd, errs := db.GetDdrDb().RetrieveWorkoutDataByPlayerCodeInDateRange(playerDetails.Code, startDate, endDate)
	if utilities.PrintErrors("failed to retrieve playcounts:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retwd_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	type profile struct {
		Name string
		Id int
		WorkoutData []ddr_models.WorkoutData
	}

	p := profile{
		Name:        playerDetails.Name,
		Id:          playerDetails.Code,
		WorkoutData: wd,
	}

	bytes, _ := json.Marshal(p)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

func ProfileWorkoutDataGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	playerDetails, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if utilities.PrintErrors("failed to retrieve player details by eagate user:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retpi_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	query := r.URL.Query()

	tz, _ := time.LoadLocation("UTC")

	startDateString := query.Get("start")
	if len(startDateString) == 0 {
		status := utilities.WriteStatus("bad","missing start param")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	endDateString := query.Get("end")
	if len(startDateString) == 0 {
		status := utilities.WriteStatus("bad","missing end param")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	start, err := time.ParseInLocation("2006-01-02", startDateString, tz)
	if err != nil {
		status := utilities.WriteStatus("bad","start must be in format: YYYY-MM-DD")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	end, err := time.ParseInLocation("2006-01-02", endDateString, tz)
	if err != nil {
		status := utilities.WriteStatus("bad","start must be in format: YYYY-MM-DD")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	workoutData, errs := db.GetDdrDb().RetrieveWorkoutDataByPlayerCodeInDateRange(playerDetails.Code, start, end)
	if utilities.PrintErrors("could not retrieve workout data:", errs) {
		status := utilities.WriteStatus("bad","ddr_retwd_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(workoutData)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}


// ProfileUpdatePatch will check the past 50 plays for the user.
// These scores will be added to the database, and then the
// difficulty details will be updated for the user. This should
// be used in favour of ProfileRefreshPatch where possible.
func ProfileUpdatePatch(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	for _, user := range users {
		client, errMsg, err := common.CreateClientForUser(user)
		if err != nil {
			status := utilities.WriteStatus("bad", errMsg)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			_, _ = rw.Write(bytes)
			return
		}

		errMsg, err = updatePlayerProfile(user, client)
		if err != nil {
			status := utilities.WriteStatus("bad", errMsg)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			_, _ = rw.Write(bytes)
			return
		}
	}

	status := utilities.WriteStatus("ok", "profile updated")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

// SongsGet will retrieve a list of all songs from the database.
// Data returned will not include the jacket image, which should
// be retrieved with the `/ddr/songs/images` endpoint.
func SongsGet(rw http.ResponseWriter, r *http.Request) {
	songIds, errs := db.GetDdrDb().RetrieveSongIds()
	if utilities.PrintErrors("failed to retrieve song ids from db:", errs) {
		status := utilities.WriteStatus("bad", "ddr_songid_db_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	query := r.URL.Query()

	var songs []ddr_models.Song
	songs, errs = db.GetDdrDb().RetrieveSongsById(songIds, query["ordering"])

	if utilities.PrintErrors("failed to retrieve songs by id:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, err := json.Marshal(songs)
	if err != nil {
		status := utilities.WriteStatus("bad", "ddr_retsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

// SongsPatch will attempt to find which songs on eagate are not
// yet in the database, and proceed to add songs as their difficulties
// to the database.
// TODO: if this fails after adding the songs to the database, new
// difficulties will be missing with no current recovery method.
func SongsPatch(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	client, errMsg, err := common.CreateClientForUser(users[0])
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	newSongs, errMsg, err := checkForNewSongs(client)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	songData, err := ddr.SongDataForClient(client, newSongs)
	if err != nil {
		status := utilities.WriteStatus("bad", "ddr_songdata_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	errs := db.GetDdrDb().AddSongs(songData)
	if utilities.PrintErrors("failed to add songs to db:", errs) {
		status := utilities.WriteStatus("bad", "ddr_addsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	songDifficulties, err := ddr.SongDifficultiesForClient(client, newSongs)
	if err != nil {
		status := utilities.WriteStatus("bad", "ddr_retdiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	errs = db.GetDdrDb().AddDifficulties(songDifficulties)
	if utilities.PrintErrors("failed to add difficulties to db:", errs) {
		status := utilities.WriteStatus("bad", "ddr_adddiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	status := utilities.WriteStatus("ok", fmt.Sprintf("added %d new songs (%d new difficulties)", len(newSongs), len(songDifficulties)))
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

func SongsReloadPatch(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	client, errMsg, err := common.CreateClientForUser(users[0])
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	songIds, err := ddr.SongIdsForClient(client)
	if err != nil {
		status := utilities.WriteStatus("bad", "err_songid_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	songData, err := ddr.SongDataForClient(client, songIds)
	if err != nil {
		status := utilities.WriteStatus("bad", "err_songdata_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	errs := db.GetDdrDb().AddSongs(songData)
	if utilities.PrintErrors("failed to add songs to db:", errs) {
		status := utilities.WriteStatus("bad", "ddr_addsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	songDifficulties, err := ddr.SongDifficultiesForClient(client, songIds)
	if err != nil {
		status := utilities.WriteStatus("bad", "ddr_diff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	errs = db.GetDdrDb().AddDifficulties(songDifficulties)
	if utilities.PrintErrors("failed to add difficulties to db:", errs) {
		status := utilities.WriteStatus("bad", "ddr_adddiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	status := utilities.WriteStatus("ok", fmt.Sprintf("added %d new songs (%d new difficulties)", len(songIds), len(songDifficulties)))
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

// SongScoresGet will retrieve score details for the user defined
// within the request JWT.
func SongsScoresGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	ddrProfile, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if utilities.PrintErrors("failed to retrieve player details for user:", errs) {
		status := utilities.WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	ids := make([]string, 0)
	if err == nil {
		var jsonMap map[string]interface{}
		err = json.Unmarshal(body, &jsonMap)

		if jsonMap["ids"] != nil {
			interfaceIds := jsonMap["ids"].([]interface{})
			for _, id := range interfaceIds {
				ids = append(ids, fmt.Sprint(id))
			}
		}
	}

	var scores []ddr_models.SongStatistics
	scores, errs = db.GetDdrDb().RetrieveSongStatisticsByPlayerCode(ddrProfile.Code, ids)

	if utilities.PrintErrors("failed to retrieve song statistics for player:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retsongstat_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(scores)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

func SongScoresGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	ddrProfile, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if utilities.PrintErrors("failed to retrieve player details for user:", errs) {
		status := utilities.WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	query := r.URL.Query()

	var scores []ddr_models.Score
	scores, errs = db.GetDdrDb().RetrieveSongScores(ddrProfile.Code, query.Get("id"), query.Get("mode"), query.Get("difficulty"), query["ordering"])

	if utilities.PrintErrors("failed to retrieve scores details for user:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retsongscore_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(scores)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

func SongJacketGet(rw http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	jacket, errs := db.GetDdrDb().RetrieveJacketForSongId(query.Get("id"))

	if utilities.PrintErrors("failed to retrieve jacket:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retsongjack_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(jacket)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

// SongScoresGet will retrieve score details for the user defined
// within the request JWT.
func SongsScoresExtendedGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	ddrProfile, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if utilities.PrintErrors("failed to retrieve player details:", errs) {
		status := utilities.WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}
	if ddrProfile.Code == 0 {
		status := utilities.WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	stats, errs := db.GetDdrDb().RetrieveExtendedScoreStatisticsByPlayerCode(ddrProfile.Code)

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(stats))
	return
}