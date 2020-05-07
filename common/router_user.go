package common

import (
	"github.com/chris-sg/bst_api/utilities"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
)

func AttachGeneralRoutes(r *mux.Router) {
	r.Path("/status").Handler(negroni.New(
		negroni.Wrap(http.HandlerFunc(Status)))).Methods(http.MethodGet)

}

// CreateUserRouter will generate a new subrouter prefixed with `/user`.
// This intends to be used for anything relating to an external user eg.
// eagate.
func CreateUserRouter() *mux.Router {
	userRouter := mux.NewRouter().PathPrefix("/user").Subrouter()

	userRouter.Path("/login").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(LoginGet)))).Methods(http.MethodGet)
	userRouter.Path("/login").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(LoginPost)))).Methods(http.MethodPost)
	userRouter.Path("/logout").Handler(utilities.GetProtectionMiddleware().With(
		negroni.Wrap(http.HandlerFunc(LogoutPost)))).Methods(http.MethodPost)

	return userRouter
}
