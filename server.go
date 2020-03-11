package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/db_builder"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"time"

	"github.com/chris-sg/bst_server_models"
)

var (
	commonMiddleware *negroni.Negroni
	protectionMiddleware *negroni.Negroni
)

func main() {
	LoadConfig()

	if dbMigration {
		db, _ := eagate_db.GetDb()
		db_builder.Create(db)
		return
	}

	r := CreateApiRouter()

	var certManager *autocert.Manager

	certManager = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(serveHost),
		Cache: autocert.DirCache("./cert_cache_api"),
	}

	srv := &http.Server{
		Handler:           r,
		Addr:		":" + servePort,
		ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	fmt.Println("api started")
	log.Fatal(srv.ListenAndServeTLS("", ""))
}

func CreateApiRouter() (r *mux.Router) {
	r = mux.NewRouter()
	apiRouter := mux.NewRouter()

	logger := negroni.NewLogger()

	commonMiddleware = negroni.New(
		negroni.HandlerFunc(logger.ServeHTTP),
		negroni.HandlerFunc(SetContentType))

	protectionMiddleware = negroni.New(
		negroni.HandlerFunc(SetForbidden),
		negroni.HandlerFunc(GetJWTMiddleware().HandlerWithNext))

	apiRouter.PathPrefix("/user").Handler(negroni.New(
		negroni.Wrap(CreateUserRouter())))

	apiRouter.PathPrefix("/ddr").Handler(negroni.New(
		negroni.Wrap(CreateDdrRouter())))

	AttachGeneralRoutes(r)

	r.PathPrefix(apiBase).Handler(commonMiddleware.With(
		negroni.Wrap(apiRouter),
	))

	return
}

var (
	nextUpdate time.Time
	cachedGate bool
	cachedDb bool
)

// Status will return any status details for the API. This
// currently caches eagate and db connections, however these
// are still regenerated every page load.
// TODO: use proper caching with timed expiry.
func Status(rw http.ResponseWriter, r *http.Request) {
	if nextUpdate.Unix() < time.Now().Unix() {
		nextUpdate = time.Now().Add(time.Minute * 2)
		updateCachedDb()
		updateCachedGate()
	}

	status := bst_models.ApiStatus{
		Api: "ok",
	}
	if cachedGate {
		status.EaGate = "ok"
	} else {
		status.EaGate = "bad"
	}
	if cachedDb {
		status.Db = "ok"
	} else {
		status.Db = "bad"
	}

	statusBytes, _ := json.Marshal(status)

	rw.WriteHeader(http.StatusOK)
	rw.Write(statusBytes)
}

// updateCachedDb will retrieve the current database status.
// This allows us to confirm whether the connection has broken
// or not.
func updateCachedDb() {
	db, err := eagate_db.GetDb()
	if err != nil || db.DB().Ping() != nil {
		cachedDb = false
	} else {
		cachedDb = true
	}
}

// updateCachedGate will get the current state of eagate. This
// will return false if either a connection cannot be formed
// with eagate or maintenance mode is active.
func updateCachedGate() {
	client := util.GenerateClient()
	cachedGate = !util.IsMaintenanceMode(client)
}

// SetForbidden will set the status header. This is done prior
// to validating a token, and will be changed if successfully
// validated.
func SetForbidden(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	if res.Status() == 0 {
		rw.WriteHeader(http.StatusForbidden)
	}
}

// SetContentType will set the content-type to json, as all api
// endpoints will return json data.
func SetContentType(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	next(rw, r)
}