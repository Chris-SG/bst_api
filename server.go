package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"time"

	"github.com/chris-sg/bst_server_models/bst_api_models"
)

var (
	protectionMiddleware *negroni.Negroni
)

func main() {
	LoadConfig()

	r := mux.NewRouter()
	apiRouter := mux.NewRouter()

	logger := negroni.NewLogger()

	commonMiddleware := negroni.New(
		negroni.HandlerFunc(logger.ServeHTTP))

	protectionMiddleware = negroni.New(
		negroni.HandlerFunc(SetForbidden),
		negroni.HandlerFunc(GetJWTMiddleware().HandlerWithNext))

	apiRouter.HandleFunc("/status", Status).Methods(http.MethodGet)
	apiRouter.PathPrefix("/user").Handler(negroni.New(
		negroni.Wrap(CreateUserRouter())))

	apiRouter.PathPrefix("/ddr").Handler(negroni.New(
		negroni.Wrap(CreateDdrRouter())))

	r.PathPrefix(apiBase).Handler(commonMiddleware.With(
		negroni.Wrap(apiRouter),
		))


	var certManager *autocert.Manager

	certManager = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(serveHost),
		Cache: autocert.DirCache("./cert_cache_api"),
	}

	srv := &http.Server{
		Handler:           r,
		Addr:		":" + servePort,
		ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	fmt.Println("api started")
	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func Status(rw http.ResponseWriter, r *http.Request) {
	status := bst_api_models.Status{
		Status: "ok",
	}

	statusBytes, _ := json.Marshal(status)

	rw.Write(statusBytes)
}

func SetForbidden(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.WriteHeader(http.StatusForbidden)
	next(rw, r)
}