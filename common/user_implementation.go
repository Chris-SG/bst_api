package common

import (
	"fmt"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/user_models"
	"github.com/chris-sg/bst_api/utilities"
	"net/http"
	"strings"
)

// tryGetEagateUsers will attempt to load any eagate users linked to
// the auth0 account provided in the request.
func TryGetEagateUsers(r *http.Request) (models []user_models.User, errMsg string, err error) {
	tokenMap := utilities.ProfileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		errMsg = "jwt_err"
		err = fmt.Errorf("failed to extract auth0 name")
		return
	}
	val = strings.ToLower(val)

	model, errs := db.GetUserDb().RetrieveUserByWebId(val)
	if utilities.PrintErrors("failed to retrieve user:", errs) {
		errMsg = "no_user"
		err = fmt.Errorf("failed to retrieve user for %s", val)
		return
	}

	if len(model.Name) == 0 {
		errMsg = "no_user"
		err = fmt.Errorf("could not find any eagate users for web id %s", val)
		return
	}
	models = append(models, model)
	return
}

// createClientForUser will generate a http client for the provided user
// model. This is intended to only be used for this specific user model,
// as it will use cookies from the database for eagate integration.
func CreateClientForUser(userModel user_models.User) (client util.EaClient, errMsg string, err error) {
	client = util.GenerateClient()
	client.SetUsername(userModel.Name)
	cookie, errs := db.GetUserDb().RetrieveUserCookieStringByUserId(userModel.Name)
	if utilities.PrintErrors("failed to retrieve cookie:", errs) {
		errMsg = "no_cookie"
		err = fmt.Errorf("failed to retrieve cookie for user %s", userModel.Name)
		return
	}
	if len(cookie) == 0 {
		errMsg = "no_cookie"
		err = fmt.Errorf("user not logged in - no cookie")
		return
	}
	client.SetEaCookie(util.CookieFromRawCookie(cookie))
	if !client.LoginState() {
		errMsg = "bad_cookie"
		err = fmt.Errorf("user not logged in - eagate rejection")
	}
	return
}