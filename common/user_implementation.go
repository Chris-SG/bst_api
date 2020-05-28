package common

import (
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"net/http"
	"strings"
)

// tryGetEagateUsers will attempt to load any eagate users linked to
// the auth0 account provided in the request.
func RetrieveEaGateUsernamesForRequest(r *http.Request) (usernames []string, err bst_models.Error) {
	err = bst_models.ErrorOK
	tokenMap := utilities.ProfileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		err = bst_models.ErrorJwtProfile
		return
	}
	val = strings.ToLower(val)

	usernames, errs := db.GetUserDb().RetrieveUsernamesByWebId(val)
	if utilities.PrintErrors("failed to retrieve user:", errs) {
		err = bst_models.ErrorNoEaUser
		return
	}

	if len(usernames) == 0 {
		err = bst_models.ErrorNoEaUser
	}
	return
}
