package main

import (
	"encoding/json"
	"fmt"
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
	tokenMap := profileFromToken(r)

	val, ok := tokenMap["name"].(string)
	if !ok {
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		panic(err)
	}

	users := user_db.RetrieveUserByWebId(db, val)

	bytes, err := json.Marshal(users)
	if err != nil {
		panic(err)
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

func LoginPost(rw http.ResponseWriter, r *http.Request) {
	tokenMap := profileFromToken(r)

	val, ok := tokenMap["name"].(string)
	if !ok {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	loginRequest := bst_web_models.LoginRequest{}
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		panic(err)
	}

	client := util.GenerateClient()

	cookie, err := user.GetCookieFromEaGate(loginRequest.Username, loginRequest.Password, client)
	if err != nil {
		rw.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		panic(err)
	}

	user_db.SetCookieForUser(db, loginRequest.Username, cookie)
	user_db.SetWebUserForUser(db, loginRequest.Username, val)

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{"status":"ok"}`))
}