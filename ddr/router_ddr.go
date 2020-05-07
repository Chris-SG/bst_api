package ddr

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/utilities"
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

	ddrRouter.Path("/profile").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfileGet)))).Methods(http.MethodGet)

	ddrRouter.Path("/profile/update").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfileUpdatePatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/profile/refresh").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(ProfileRefreshPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs", SongsGet).Methods(http.MethodGet)
	ddrRouter.Path("/songs").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongsPatch)))).Methods(http.MethodPatch)

	ddrRouter.Path("/songsreload").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(SongsReloadPatch)))).Methods(http.MethodPatch)

	ddrRouter.HandleFunc("/songs/jackets", SongsJacketGet).Methods(http.MethodGet)

	ddrRouter.HandleFunc("/songs/{id:[A-Za-z0-9]{32}}", SongsIdGet).Methods(http.MethodGet)

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
	playerDetails, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if utilities.PrintErrors("failed to retrieve player details by eagate user:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retpi_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	wd, errs := eagate_db.GetDdrDb().RetrieveWorkoutDataByPlayerCodeInDateRange(playerDetails.Code, 29, 0)
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
	ordering := ""

	body, err := ioutil.ReadAll(r.Body)
	if err == nil {
		orderStruct := bst_models.Ordering{}
		err := json.Unmarshal(body, &orderStruct)
		if err == nil {
			ordering = utilities.ValidateOrdering(ddr_models.Song{}, orderStruct.OrderBy)
		}
	}

	songIds, errs := eagate_db.GetDdrDb().RetrieveSongIds()
	if utilities.PrintErrors("failed to retrieve song ids from db:", errs) {
		status := utilities.WriteStatus("bad", "ddr_songid_db_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	var songs []ddr_models.Song

	if ordering == "" {
		songs, errs = eagate_db.GetDdrDb().RetrieveSongsById(songIds)
	} else {
		songs, errs = eagate_db.GetDdrDb().RetrieveOrderedSongsById(songIds, ordering)
	}
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
	errs := eagate_db.GetDdrDb().AddSongs(songData)
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
	errs = eagate_db.GetDdrDb().AddDifficulties(songDifficulties)
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
	errs := eagate_db.GetDdrDb().AddSongs(songData)
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
	errs = eagate_db.GetDdrDb().AddDifficulties(songDifficulties)
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

// SongJacketsGet will retrieve the jackets for song ids provided in the
// request body.
func SongsJacketGet(rw http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status := utilities.WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	ids := bst_models.DdrSongIds{}
	err = json.Unmarshal(body, &ids)
	if err != nil {
		status := utilities.WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	jackets, errs := eagate_db.GetDdrDb().RetrieveJacketsForSongIds(ids.Ids)
	if utilities.PrintErrors("failed to retrieve jackets for song ids:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retjack_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
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
	_, _ = rw.Write(bytes)
	return
}

// SongsIdGet will retrieve details about the song id located within
// the request URI.
// TODO improve data returned
func SongsIdGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val := vars["id"]

	song, errs := eagate_db.GetDdrDb().RetrieveSongByIdWithJacket(val)
	if utilities.PrintErrors("failed to retrieve song id:", errs) {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(song.Id) == 0 {
		status := utilities.WriteStatus("bad", fmt.Sprintf("unable to find song id %s in database", val))
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	difficulties, errs := eagate_db.GetDdrDb().RetrieveDifficultiesById([]string{val})
	if utilities.PrintErrors("failed to retrieve difficulties:", errs) {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(difficulties) == 0 {
		status := utilities.WriteStatus("bad", "ddr_retdiff_fail")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(difficulties)
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

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
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

	if len(ids) > 0 {
		scores, errs = eagate_db.GetDdrDb().RetrieveSongStatisticsByPlayerCodeForSongIds(ddrProfile.Code, ids)
	} else {
		eagate_db.GetDdrDb().RetrieveSongStatisticsByPlayerCode(ddrProfile.Code)
	}
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

func SongsScoresIdDiffGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	mode := strings.ToUpper(vars["mode"])
	diff := strings.ToUpper(vars["difficulty"])
	mode = utilities.CleanString(mode)
	diff = utilities.CleanString(diff)

	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
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
	scores, errs := eagate_db.GetDdrDb().RetrieveScoresByPlayerCodeForChart(ddrProfile.Code, id, mode, diff)
	if utilities.PrintErrors("failed to retrieve scores for player:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retsongstat_err")
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

func SongsScoresIdGet(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	val := vars["id"]

	users, errMsg, err := common.TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
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

	result := bst_models.DdrScoresDetailed{ Id: val }

	statistics, errs := eagate_db.GetDdrDb().RetrieveSongStatisticsByPlayerCodeForSongIds(ddrProfile.Code, []string{val})
	if utilities.PrintErrors("failed to retrieve song statistics from db:", errs) {
		status := utilities.WriteStatus("bad", "ddr_retsongstat_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		_, _ = rw.Write(bytes)
		return
	}

	bytes, _ := json.Marshal(statistics)
	_ = json.Unmarshal(bytes, &result.TopScores)

	mode := map[string][]bst_models.DdrScoresDetailedDifficulty{}
	scoresMap := map[string][]bst_models.DdrScoresDetailedScore{}
	for _, stat := range statistics {
		scores, errs := eagate_db.GetDdrDb().RetrieveScoresByPlayerCodeForChart(ddrProfile.Code, val, stat.Mode, stat.Difficulty)
		if utilities.PrintErrors("failed to retrieve scores for player:", errs) {
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

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
	if utilities.PrintErrors("failed to retrieve player details for user:", errs) {
		status := utilities.WriteStatus("bad", "no_user")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = rw.Write(bytes)
		return
	}

	query := r.URL.Query()

	var scores []ddr_models.Score
	scores, errs = eagate_db.GetDdrDb().RetrieveSongScores(ddrProfile.Code, query.Get("id"), query.Get("mode"), query.Get("difficulty"), query["ordering"])

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

	jacket, errs := eagate_db.GetDdrDb().RetrieveJacketForSongId(query.Get("id"))

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

	ddrProfile, errs := eagate_db.GetDdrDb().RetrievePlayerDetailsByEaGateUser(users[0].Name)
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

	stats, errs := eagate_db.GetDdrDb().RetrieveExtendedScoreStatisticsByPlayerCode(ddrProfile.Code)

	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(stats))
	return
}