package main

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_server_models"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/chris-sg/eagate/user"
)

// CreateUserRouter will generate a new subrouter prefixed with `/user`.
// This intends to be used for anything relating to an external user eg.
// eagate.
func CreateUserRouter() *mux.Router {
	userRouter := mux.NewRouter().PathPrefix("/user").Subrouter()

	userRouter.Path("/login").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(LoginGet)))).Methods(http.MethodGet)
	userRouter.Path("/login").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(LoginPost)))).Methods(http.MethodPost)
	userRouter.Path("/logout").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(LogoutPost)))).Methods(http.MethodPost)

	return userRouter
}

// LoginGet will retrieve any relations between the requester and the
// database. This may produce multiple relations in the case a user
// has linked multiple accounts. Any stored cookies will be nullified.
func LoginGet(rw http.ResponseWriter, r *http.Request) {
	users, errMsg, err := tryGetEagateUsers(r)
	if err != nil {
		status := WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	eagateUsers := make([]bst_models.EagateUser, 0)

	for i, _ := range users {
		eagateUser := bst_models.EagateUser{
			Username: users[i].Name,
			Expired:  users[i].Expiration < time.Now().UnixNano()/1000,
		}
		if !eagateUser.Expired {
			client, _, err := createClientForUser(users[i])
			if err != nil || !client.LoginState() {
				fmt.Println(err)
				eagateUser.Expired = true
			}
		}
		eagateUsers = append(eagateUsers, eagateUser)
	}

	bytes, err := json.Marshal(eagateUsers)
	if err != nil {
		status := WriteStatus("bad", "marshal_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// LoginPost will attempt to login to eagate using the supplied credentials.
// This will not check the current user state, which if exists, will have
// its cookie updated and web user replaced (if applicable).
func LoginPost(rw http.ResponseWriter, r *http.Request) {
	tokenMap := profileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		status := WriteStatus("bad", "jwt_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}
	val = strings.ToLower(val)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("%s\n", err.Error())
		status := WriteStatus("bad", "bad_api_req")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	loginRequest := bst_models.LoginRequest{}
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		glog.Warningf("failed to decode login request for %s: %s\n", loginRequest.Username, err.Error())
		status := WriteStatus("bad", "unmarshal_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	glog.Infof("user %s attempting to login to eagate\n", loginRequest.Username)
	client := util.GenerateClient()

	cookie, err := user.GetCookieFromEaGate(loginRequest.Username, loginRequest.Password, loginRequest.OneTimePassword, client)
	if err != nil {
		status := WriteStatus("bad", "bad_cookie")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	errs := eagate_db.GetUserDb().SetCookieForUser(loginRequest.Username, cookie)
	if len(errs) > 0 {
		status := WriteErrorStatus(errs)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}
	errs = eagate_db.GetUserDb().SetWebUserForEaUser(loginRequest.Username, val)
	if len(errs) > 0 {
		status := WriteErrorStatus(errs)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	status := WriteStatus("ok", "loaded cookie for user")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
	return
}

func LogoutPost(rw http.ResponseWriter, r *http.Request) {
	tokenMap := profileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		status := WriteStatus("bad", "jwt_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}
	val = strings.ToLower(val)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status := WriteStatus("bad", "bad_api_req")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	logoutRequest := bst_models.LogoutRequest{}
	err = json.Unmarshal(body, &logoutRequest)
	if err != nil {
		status := WriteStatus("bad", "unmarshal_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	user, errs := eagate_db.GetUserDb().RetrieveUserByUserId(logoutRequest.Username)
	if len(errs) > 0 {
		status := WriteErrorStatus(errs)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	if user.WebUser == val {
		errs := eagate_db.GetUserDb().SetWebUserForEaUser(user.Name, "")
		if len(errs) > 0 {
			status := WriteErrorStatus(errs)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(bytes)
			return
		}

		status := WriteStatus("ok", "user unlinked")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusOK)
		rw.Write(bytes)
		return
	}

	status := WriteStatus("bad", "no_user")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusUnauthorized)
	rw.Write(bytes)
	return
}