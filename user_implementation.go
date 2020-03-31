package main

import (
	"fmt"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_models/user_models"
	"net/http"
	"strings"
)

// tryGetEagateUsers will attempt to load any eagate users linked to
// the auth0 account provided in the request.
func tryGetEagateUsers(r *http.Request) (models []user_models.User, err error) {
	tokenMap := profileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		err = fmt.Errorf("failed to extract auth0 name")
		return
	}
	val = strings.ToLower(val)

	model, errs := eagate_db.GetUserDb().RetrieveUserByWebId(val)
	if PrintErrors("failed to retrieve user:", errs) {
		err = fmt.Errorf("failed to retrieve user for %s", val)
		return
	}

	if len(model.Name) == 0 {
		err = fmt.Errorf("could not find any eagate users for web id %s", val)
		return
	}
	models = append(models, model)
	return
}

// createClientForUser will generate a http client for the provided user
// model. This is intended to only be used for this specific user model,
// as it will use cookies from the database for eagate integration.
func createClientForUser(userModel user_models.User) (client util.EaClient, err error) {
	client = util.GenerateClient()
	client.SetUsername(userModel.Name)
	cookie, errs := eagate_db.GetUserDb().RetrieveUserCookieStringByUserId(userModel.Name)
	if PrintErrors("failed to retrieve cookie:", errs) {
		err = fmt.Errorf("failed to retrieve cookie for user %s", userModel.Name)
		return
	}
	if len(cookie) == 0 {
		err = fmt.Errorf("user not logged in - no cookie")
		return
	}
	client.SetEaCookie(util.CookieFromRawCookie(cookie))
	if !client.LoginState() {
		err = fmt.Errorf("user not logged in - eagate rejection")
	}
	return
}