package utilities

import (
	"github.com/urfave/negroni"
	"net/http"
)

var (
	commonMiddleware *negroni.Negroni
	protectionMiddleware *negroni.Negroni
)

func PrepareMiddleware() {
	logger := negroni.NewLogger()

	if commonMiddleware == nil {
		commonMiddleware = negroni.New(
			negroni.HandlerFunc(logger.ServeHTTP),
			negroni.HandlerFunc(setContentType))
	}

	if protectionMiddleware == nil {
		protectionMiddleware = negroni.New(
			negroni.HandlerFunc(setForbidden),
			negroni.HandlerFunc(GetJWTMiddleware().HandlerWithNext))
	}
}

func GetCommonMiddleware() (*negroni.Negroni) {
	return commonMiddleware
}

func GetProtectionMiddleware() (*negroni.Negroni) {
	return protectionMiddleware
}

// SetForbidden will set the status header. This is done prior
// to validating a token, and will be changed if successfully
// validated.
func setForbidden(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	if res.Status() == 0 {
		rw.Write([]byte("bad_token"))
		rw.WriteHeader(http.StatusForbidden)
	}
}

// SetContentType will set the content-type to json, as all api
// endpoints will return json data.
func setContentType(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	next(rw, r)
}