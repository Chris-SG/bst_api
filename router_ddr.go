package main

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_server_models"
	"github.com/chris-sg/eagate/ddr"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_models/ddr_models"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
	"strings"
)

// CreateDdrRouter will create a mux router to be attached to
// the main router, prefixed with '/ddr'.
func CreateDdrRouter() *mux.Router {
	ddrRouter := mux.NewRouter().PathPrefix("/ddr").Subrouter()

	ddrRouter.Path("/profile").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(ProfileGet)))).Methods(http.MethodGet)

	ddrRouter.Path("/profile/update").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(ProfileUpdatePatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/profile/refresh").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(ProfileRefreshPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs", SongsGet).Methods(http.MethodGet)
	ddrRouter.Path("/songs").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsPatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/songsreload").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsReloadPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs/jackets", SongsJacketGet).Methods(http.MethodGet)

	ddrRouter.HandleFunc("/songs/{id:[A-Za-z0-9]{32}}", SongsIdGet).Methods(http.MethodGet)

	ddrRouter.Path("/songs/scores").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsScoresGet)))).Methods(http.MethodGet)
	ddrRouter.Path("/songs/scores/{id:[A-Za-z0-9]{32}}/{mode:[A-Za-z]+}/{difficulty:[A-Za-z]+}").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsScoresIdDiffGet)))).Methods(http.MethodGet)
	ddrRouter.Path("/songs/scores/{id:[A-Za-z0-9]{32}}").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsScoresIdGet)))).Methods(http.MethodGet)

	ddrRouter.Path("/songs/scores/extended").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(SongsScoresExtendedGet)))).Methods(http.MethodGet)

	return ddrRouter
}

// ProfileRefreshPatch will perform a full refresh of all song
// difficulties in the database for the user. This is an expensive
// operation and should be used sparingly.
func ProfileRefreshPatch(rw http.ResponseWriter, r *http.Request) {
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

		errMsg, err = refreshDdrUser(client)
		if err != nil {
			status := WriteStatus("bad", errMsg)
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

// ProfileGet will retrieve formatted profile details for
// the current user.
func ProfileGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	playerDetails, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if PrintErrors("failed to retrieve player details by eagate user:", errs) {
		status := WriteStatus("bad", "ddr_retpi_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	wd, errs := eagate_db.GetDdrDb().RetrieveWorkoutDataByPlayerCodeInDateRange(playerDetails.Code, 29, 0)
	if PrintErrors("failed to retrieve playcounts:", errs) {
		status := WriteStatus("bad", "ddr_retwd_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
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
	rw.Write(bytes)
	return
}

// ProfileUpdatePatch will check the past 50 plays for the user.
// These scores will be added to the database, and then the
// difficulty details will be updated for the user. This should
// be used in favour of ProfileRefreshPatch where possible.
func ProfileUpdatePatch(rw http.ResponseWriter, r *http.Request) {
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

		errMsg, err = updatePlayerProfile(user, client)
		if err != nil {
			status := WriteStatus("bad", errMsg)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(bytes)
			return
		}
	}

	status := WriteStatus("ok", "profile updated")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongsGet will retrieve a list of all songs from the database.
// Data returned will not include the jacket image, which should
// be retrieved with the `/ddr/songs/images` endpoint.
func SongsGet(rw http.ResponseWriter, r *http.Request) {
	ordering := ""

	body, err := ioutil.ReadAll(r.Body)
	if err == nil {
		orderStruct := bst_models.Ordering{}
		err := json.Unmarshal(body, &orderStruct)
		if err == nil {
			ordering = ValidateOrdering(ddr_models.Song{}, orderStruct.OrderBy)
		}
	}

	songIds, errs := eagate_db.GetDdrDb().RetrieveSongIds()
	if PrintErrors("failed to retrieve song ids from db:", errs) {
		status := WriteStatus("bad", "ddr_songid_db_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	var songs []ddr_models.Song

	if ordering == "" {
		songs, errs = eagate_db.GetDdrDb().RetrieveSongsById(songIds)
	} else {
		songs, errs = eagate_db.GetDdrDb().RetrieveOrderedSongsById(songIds, ordering)
	}
	if PrintErrors("failed to retrieve songs by id:", errs) {
		status := WriteStatus("bad", "ddr_retsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	bytes, err := json.Marshal(songs)
	if err != nil {
		status := WriteStatus("bad", "ddr_retsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongsPatch will attempt to find which songs on eagate are not
// yet in the database, and proceed to add songs as their difficulties
// to the database.
// TODO: if this fails after adding the songs to the database, new
// difficulties will be missing with no current recovery method.
func SongsPatch(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	client, errMsg, err := createClientForUser(users[0])
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	newSongs, errMsg, err := checkForNewSongs(client)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	songData, err := ddr.SongDataForClient(client, newSongs)
	if err != nil {
		status := WriteStatus("bad", "ddr_songdata_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	errs := eagate_db.GetDdrDb().AddSongs(songData)
	if PrintErrors("failed to add songs to db:", errs) {
		status := WriteStatus("bad", "ddr_addsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	songDifficulties, err := ddr.SongDifficultiesForClient(client, newSongs)
	if err != nil {
		status := WriteStatus("bad", "ddr_retdiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	errs = eagate_db.GetDdrDb().AddDifficulties(songDifficulties)
	if PrintErrors("failed to add difficulties to db:", errs) {
		status := WriteStatus("bad", "ddr_adddiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	status := WriteStatus("ok", fmt.Sprintf("added %d new songs (%d new difficulties)", len(newSongs), len(songDifficulties)))
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func SongsReloadPatch(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	client, errMsg, err := createClientForUser(users[0])
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	songIds, err := ddr.SongIdsForClient(client)
	if err != nil {
		status := WriteStatus("bad", "err_songid_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	songData, err := ddr.SongDataForClient(client, songIds)
	if err != nil {
		status := WriteStatus("bad", "err_songdata_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	errs := eagate_db.GetDdrDb().AddSongs(songData)
	if PrintErrors("failed to add songs to db:", errs) {
		status := WriteStatus("bad", "ddr_addsong_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	songDifficulties, err := ddr.SongDifficultiesForClient(client, songIds)
	if err != nil {
		status := WriteStatus("bad", "ddr_diff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}
	errs = eagate_db.GetDdrDb().AddDifficulties(songDifficulties)
	if PrintErrors("failed to add difficulties to db:", errs) {
		status := WriteStatus("bad", "ddr_adddiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	status := WriteStatus("ok", fmt.Sprintf("added %d new songs (%d new difficulties)", len(songIds), len(songDifficulties)))
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongJacketsGet will retrieve the jackets for song ids provided in the
// request body.
func SongsJacketGet(rw http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	ids := bst_models.DdrSongIds{}
	err = json.Unmarshal(body, &ids)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	jackets, errs := eagate_db.GetDdrDb().RetrieveJacketsForSongIds(ids.Ids)
	if PrintErrors("failed to retrieve jackets for song ids:", errs) {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}


	idWithJacket := make([]bst_models.DdrSongIdWithJacket, len(jackets))
	i := 0
	for k, v := range jackets {
		idWithJacket[i].Id = k
		idWithJacket[i].Jacket = v
		i++
	}

	bytes, _ := json.Marshal(jackets)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongsIdGet will retrieve details about the song id located within
// the request URI.
// TODO improve data returned
func SongsIdGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val := vars["id"]

	song, errs := eagate_db.GetDdrDb().RetrieveSongByIdWithJacket(val)
	if PrintErrors("failed to retrieve song id:", errs) {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(song.Id) == 0 {
		status := WriteStatus("bad", fmt.Sprintf("unable to find song id %s in database", val))
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	difficulties, errs := eagate_db.GetDdrDb().RetrieveDifficultiesById([]string{val})
	if PrintErrors("failed to retrieve difficulties:", errs) {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(difficulties) == 0 {
		status := WriteStatus("bad", "ddr_retdiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(difficulties)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongScoresGet will retrieve score details for the user defined
// within the request JWT.
func SongsScoresGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if PrintErrors("failed to retrieve player details for user:", errs) {
		status := WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
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

	if len(ids) > 0 {
		scores, errs = eagate_db.GetDdrDb().RetrieveSongStatisticsByPlayerCodeForSongIds(ddrProfile.Code, ids)
	} else {
		eagate_db.GetDdrDb().RetrieveSongStatisticsByPlayerCode(ddrProfile.Code)
	}
	if PrintErrors("failed to retrieve song statistics for player:", errs) {
		status := WriteStatus("bad", "ddr_retsongstat_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(scores)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func SongsScoresIdDiffGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mode := strings.ToUpper(vars["mode"])
	diff := strings.ToUpper(vars["difficulty"])
	mode = cleanString(mode)
	diff = cleanString(diff)

	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if PrintErrors("failed to retrieve player details:", errs) {
		status := WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}
	if ddrProfile.Code == 0 {
		status := WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}
	scores, errs := eagate_db.GetDdrDb().RetrieveScoresByPlayerCodeForChart(ddrProfile.Code, id, mode, diff)
	if PrintErrors("failed to retrieve scores for player:", errs) {
		status := WriteStatus("bad", "ddr_retsongstat_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(scores)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func SongsScoresIdGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val := vars["id"]

	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if PrintErrors("failed to retrieve player details:", errs) {
		status := WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}
	if ddrProfile.Code == 0 {
		status := WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	result := bst_models.DdrScoresDetailed{ Id: val }

	statistics, errs := eagate_db.GetDdrDb().RetrieveSongStatisticsByPlayerCodeForSongIds(ddrProfile.Code, []string{val})
	if PrintErrors("failed to retrieve song statistics from db:", errs) {
		status := WriteStatus("bad", "ddr_retsongstat_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(statistics)
	json.Unmarshal(bytes, &result.TopScores)

	mode := map[string][]bst_models.DdrScoresDetailedDifficulty{}
	scoresMap := map[string][]bst_models.DdrScoresDetailedScore{}
	for _, stat := range statistics {
		scores, errs := eagate_db.GetDdrDb().RetrieveScoresByPlayerCodeForChart(ddrProfile.Code, val, stat.Mode, stat.Difficulty)
		if PrintErrors("failed to retrieve scores for player:", errs) {
			continue
		}

		for _, score := range scores {
			scoresMap[stat.Mode + ":" + stat.Difficulty] = append(scoresMap[stat.Mode + ":" + stat.Difficulty], bst_models.DdrScoresDetailedScore{
				Score:       score.Score,
				ClearStatus: score.ClearStatus,
				TimePlayed:  score.TimePlayed,
			})
		}
	}

	for k, v := range scoresMap {
		s := strings.Split(k, ":")
		d := bst_models.DdrScoresDetailedDifficulty{
			Difficulty: s[1],
			Scores:     v,
		}
		mode[s[0]] = append(mode[s[0]], d)
	}

	for k, v := range mode {
		result.Modes = append(result.Modes, bst_models.DdrScoresDetailedMode{
			Mode:         k,
			Difficulties: v,
		})
	}

	bytes, _ = json.Marshal(result)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// SongScoresGet will retrieve score details for the user defined
// within the request JWT.
func SongsScoresExtendedGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if PrintErrors("failed to retrieve player details:", errs) {
		status := WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}
	if ddrProfile.Code == 0 {
		status := WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	stats, errs := eagate_db.GetDdrDb().RetrieveExtendedScoreStatisticsByPlayerCode(ddrProfile.Code)

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(stats))
}