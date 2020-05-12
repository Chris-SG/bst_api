package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/common"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/ddr"
	"github.com/chris-sg/bst_api/drs"
	"github.com/chris-sg/bst_api/utilities"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	utilities.LoadConfig()
	utilities.PrepareMiddleware()

	if utilities.DbMigration {
		db.GetMigrator().Create()
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

	apiRouter.Path("/runmigration").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(RunDbMigration)))).Methods(http.MethodPatch)

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

func RunDbMigration(rw http.ResponseWriter, r *http.Request) {
	requiredScopes := []string{"update:database"}
	tokenMap := utilities.ProfileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		status := utilities.WriteStatus("bad", "jwt_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}
	val = strings.ToLower(val)
	if !utilities.UserHasScopes(val, requiredScopes) {
		glog.Warningf(
			"user %s tried to migrate db, but did not have required scopes %s",
			val,
			strings.Join(requiredScopes, ","))
		status := utilities.WriteStatus("bad", "scope_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		_, _ = rw.Write(bytes)
		return
	}

	db.GetMigrator().Create()

	status := utilities.WriteStatus("ok", "migrated")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(bytes)
	return
}