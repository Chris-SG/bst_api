package main

import (
	"crypto/tls"
	"fmt"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/ddr"
	"github.com/chris-sg/bst_api/drs"
	"github.com/chris-sg/bst_api/utilities"
	"github.com/chris-sg/eagate_db"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"time"
)

func main() {
	utilities.LoadConfig()
	utilities.PrepareMiddleware()

	if utilities.DbMigration {
		eagate_db.GetMigrator().Create()
		return
	}

	r := CreateApiRouter()

	var certManager *autocert.Manager

	certManager = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(utilities.ServeHost),
		Cache: autocert.DirCache("./cert_cache_api"),
	}

	srv := &http.Server{
		Handler:           r,
		Addr:		":" + utilities.ServePort,
		ReadTimeout: 30 * time.Second,
		WriteTimeout: 90 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go func() {
		// serve HTTP, which will redirect automatically to HTTPS
		h := certManager.HTTPHandler(nil)
		log.Fatal(http.ListenAndServe(":http", h))
	}()

	fmt.Println("api started")
	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func CreateApiRouter() (r *mux.Router) {
	r = mux.NewRouter()
	apiRouter := mux.NewRouter()

	apiRouter.PathPrefix("/user").Handler(negroni.New(
		negroni.Wrap(common.CreateUserRouter())))

	apiRouter.PathPrefix("/ddr").Handler(negroni.New(
		negroni.Wrap(ddr.CreateDdrRouter())))

	apiRouter.PathPrefix("/drs").Handler(negroni.New(
		negroni.Wrap(drs.CreateDrsRouter())))

	common.AttachGeneralRoutes(r)

	r.PathPrefix(utilities.ApiBase).Handler(utilities.GetCommonMiddleware().With(
		negroni.Wrap(apiRouter),
	))

	return
}
