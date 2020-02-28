package main

import (
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

func CreateUserRouter() *mux.Router {
	userRouter := mux.NewRouter().PathPrefix("/user").Subrouter()

	userRouter.Path("/login").Handler(protectionMiddleware.With(
		negroni.Wrap(http.HandlerFunc(LoginPost)))).Methods(http.MethodPost)

	return userRouter
}

func LoginPost(rw http.ResponseWriter, r *http.Request) {

}