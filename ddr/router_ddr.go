package ddr

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/ddr"
	"github.com/chris-sg/bst_api/eagate/user"
	"github.com/chris-sg/bst_api/models/ddr_models"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	userModel, exists, errs := db.GetUserDb().RetrieveUserByUserId(usernames[0])
	if !exists {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	if utilities.PrintErrors("failed to retrieve user: ", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorUnknownUser)
		return
	}
	client, err := user.CreateClientForUser(userModel)
	defer client.UpdateCookie()
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	err = refreshDdrUser(client)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	utilities.RespondWithError(rw, err)
	return
}

// ProfileGet will retrieve formatted profile details for
// the current user.
func ProfileGet(rw http.ResponseWriter, r *http.Request) {
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	playerDetails, exists, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(usernames[0])
	if !exists {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerNotFound)
		return
	}
	if utilities.PrintErrors("failed to retrieve player details by eagate user:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerInfoDbRead)
		return
	}
	today := time.Now()
	tz, _ := time.LoadLocation("UTC")

	endDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, tz)
	startDate := endDate.AddDate(0, -1, 0)
	wd, errs := db.GetDdrDb().RetrieveWorkoutDataByPlayerCodeInDateRange(playerDetails.Code, startDate, endDate)
	if utilities.PrintErrors("failed to retrieve playcounts:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrStatsDbRead)
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	playerDetails, exists, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(usernames[0])
	if !exists {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerNotFound)
		return
	}
	if utilities.PrintErrors("failed to retrieve player details by eagate user:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerInfoDbRead)
		return
	}
	query := r.URL.Query()

	tz, _ := time.LoadLocation("UTC")

	startDateString := query.Get("start")
	if len(startDateString) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorBadQuery)
		return
	}

	endDateString := query.Get("end")
	if len(startDateString) == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorBadQuery)
		return
	}

	start, e := time.ParseInLocation("2006-01-02", startDateString, tz)
	if e != nil {
		utilities.RespondWithError(rw, bst_models.ErrorTimeParse)
	}

	end, e := time.ParseInLocation("2006-01-02", endDateString, tz)
	if e != nil {
		utilities.RespondWithError(rw, bst_models.ErrorTimeParse)
		return
	}

	workoutData, errs := db.GetDdrDb().RetrieveWorkoutDataByPlayerCodeInDateRange(playerDetails.Code, start, end)
	if utilities.PrintErrors("could not retrieve workout data:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrStatsDbRead)
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	userModel, exists, errs := db.GetUserDb().RetrieveUserByUserId(usernames[0])
	if !exists {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	if utilities.PrintErrors("failed to retrieve user: ", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorUnknownUser)
		return
	}
	client, err := user.CreateClientForUser(userModel)
	defer client.UpdateCookie()
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	err = UpdatePlayerProfile(userModel, client)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	utilities.RespondWithError(rw, bst_models.ErrorOK)
	return
}

// SongsGet will retrieve a list of all songs from the database.
// Data returned will not include the jacket image, which should
// be retrieved with the `/ddr/songs/images` endpoint.
func SongsGet(rw http.ResponseWriter, r *http.Request) {
	songIds, errs := db.GetDdrDb().RetrieveSongIds()
	if utilities.PrintErrors("failed to retrieve song ids from db:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrSongIdsDbRead)
		return
	}

	query := r.URL.Query()

	var songs []ddr_models.Song
	songs, errs = db.GetDdrDb().RetrieveSongsById(songIds, query["ordering"])

	if utilities.PrintErrors("failed to retrieve songs by id:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrSongDataDbRead)
		return
	}

	bytes, err := json.Marshal(songs)
	if err != nil {
		utilities.RespondWithError(rw, bst_models.ErrorJsonEncode)
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	userModel, exists, errs := db.GetUserDb().RetrieveUserByUserId(usernames[0])
	if utilities.PrintErrors("failed to retrieve user from db:", errs) || !exists {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	client, err := user.CreateClientForUser(userModel)
	defer client.UpdateCookie()
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	newSongs, err := checkForNewSongs(client)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	songData, err := ddr.SongDataForClient(client, newSongs)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	errs = db.GetDdrDb().AddSongs(songData)
	if utilities.PrintErrors("failed to add songs to db:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrSongDataDbWrite)
		return
	}

	songDifficulties, err := ddr.SongDifficultiesForClient(client, newSongs)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	errs = db.GetDdrDb().AddDifficulties(songDifficulties)
	if utilities.PrintErrors("failed to add difficulties to db:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrSongDifficultiesDbWrite)
		return
	}

	utilities.RespondWithError(rw, bst_models.ErrorOK)
	return
}

func SongsReloadPatch(rw http.ResponseWriter, r *http.Request) {
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	userModel, exists, errs := db.GetUserDb().RetrieveUserByUserId(usernames[0])
	if utilities.PrintErrors("failed to retrieve user from db:", errs) || !exists {
		utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
		return
	}
	client, err := user.CreateClientForUser(userModel)
	defer client.UpdateCookie()
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	songIds, err := ddr.SongIdsForClient(client)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	songData, err := ddr.SongDataForClient(client, songIds)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	errs = db.GetDdrDb().AddSongs(songData)
	if utilities.PrintErrors("failed to add songs to db:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrSongDataDbWrite)
		return
	}

	songDifficulties, err := ddr.SongDifficultiesForClient(client, songIds)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	errs = db.GetDdrDb().AddDifficulties(songDifficulties)
	if utilities.PrintErrors("failed to add difficulties to db:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrSongDifficultiesDbWrite)
		return
	}

	utilities.RespondWithError(rw, bst_models.ErrorOK)
	return
}

// SongScoresGet will retrieve score details for the user defined
// within the request JWT.
func SongsScoresGet(rw http.ResponseWriter, r *http.Request) {
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	ddrProfile, exists, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(usernames[0])
	if !exists {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerNotFound)
		return
	}
	if utilities.PrintErrors("failed to retrieve player details for user:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerInfoDbRead)
		return
	}

	body, e := ioutil.ReadAll(r.Body)
	ids := make([]string, 0)
	if e == nil {
		var jsonMap map[string]interface{}
		e = json.Unmarshal(body, &jsonMap)

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
		utilities.RespondWithError(rw, bst_models.ErrorDdrStatsDbRead)
		return
	}

	bytes, _ := json.Marshal(scores)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}

func SongScoresGet(rw http.ResponseWriter, r *http.Request) {
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	ddrProfile, exists, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(usernames[0])
	if !exists {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerNotFound)
		return
	}
	if utilities.PrintErrors("failed to retrieve player details for user:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerInfoDbRead)
		return
	}

	query := r.URL.Query()

	var scores []ddr_models.Score
	scores, errs = db.GetDdrDb().RetrieveSongScores(ddrProfile.Code, query.Get("id"), query.Get("mode"), query.Get("difficulty"), query["ordering"])

	if utilities.PrintErrors("failed to retrieve scores details for user:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrStatsDbRead)
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
		utilities.RespondWithError(rw, bst_models.ErrorDdrSongDataDbRead)
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
	usernames, err := common.RetrieveEaGateUsernamesForRequest(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	ddrProfile, exists, errs := db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(usernames[0])
	if !exists {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerNotFound)
		return
	}
	if utilities.PrintErrors("failed to retrieve player details:", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerInfoDbRead)
		return
	}
	if ddrProfile.Code == 0 {
		utilities.RespondWithError(rw, bst_models.ErrorDdrPlayerInfoDbRead)
		return
	}

	stats, errs := db.GetDdrDb().RetrieveExtendedScoreStatisticsByPlayerCode(ddrProfile.Code)

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(stats))
	return
}