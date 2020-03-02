package main

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_server_models/bst_api_models"
	"github.com/chris-sg/bst_server_models/bst_web_models"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/user_db"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"

	"github.com/chris-sg/eagate/user"
)

func CreateUserRouter() *mux.Router {
	userRouter := mux.NewRouter().PathPrefix("/user").Subrouter()

	userRouter.Path("/login").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(LoginGet)))).Methods(http.MethodGet)
	userRouter.Path("/login").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(LoginPost)))).Methods(http.MethodPost)

	return userRouter
}

func LoginGet(rw http.ResponseWriter, r *http.Request) {
	users, err := tryGetEagateUser(r)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	for i, _ := range users {
		users[i].Cookie = "***"
	}

	bytes, err := json.Marshal(users)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func LoginPost(rw http.ResponseWriter, r *http.Request) {
	tokenMap := profileFromToken(r)

	val, ok := tokenMap["name"].(string)
	if !ok {
		status := WriteStatus("bad", "failed to read auth name from token")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	loginRequest := bst_web_models.LoginRequest{}
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	client := util.GenerateClient()

	cookie, err := user.GetCookieFromEaGate(loginRequest.Username, loginRequest.Password, client)
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		status := WriteStatus("bad", err.Error())
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
		return
	}

	user_db.SetCookieForUser(db, loginRequest.Username, cookie)
	user_db.SetWebUserForUser(db, loginRequest.Username, val)

	status := WriteStatus("ok", "loaded cookie for user")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
	return
}