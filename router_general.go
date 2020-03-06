package main

import (
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

func AttachGeneralRoutes(r *mux.Router) {
	r.Path("/status").Handler(negroni.New(
		negroni.Wrap(http.HandlerFunc(Status)))).Methods(http.MethodGet)

}