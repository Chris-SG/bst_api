package main

import (
	"fmt"
	"github.com/chris-sg/eagate/user"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
	"github.com/chris-sg/eagate_db/user_db"
	"github.com/chris-sg/eagate_models/user_models"
	"net/http"
)

// tryGetEagateUsers will attempt to load any eagate users linked to
// the auth0 account provided in the request.
func tryGetEagateUsers(r *http.Request) (models []user_models.User, err error) {
	tokenMap := profileFromToken(r)

	val, ok := tokenMap["name"].(string)
	if !ok {
		err = fmt.Errorf("failed to extract auth0 name")
		return
	}

	db, err := eagate_db.GetDb()
	if err != nil {
		return
	}

	models = user_db.RetrieveUserByWebId(db, val)
	if len(models) == 0 {
		err = fmt.Errorf("could not find any eagate users for web id %s", val)
	}
	return
}

// createClientForUser will generate a http client for the provided user
// model. This is intended to only be used for this specific user model,
// as it will use cookies from the database for eagate integration.
func createClientForUser(userModel user_models.User) (client util.EaClient, err error) {
	client = util.GenerateClient()
	db, err := eagate_db.GetDb()
	if err != nil {
		return
	}
	cookie := user_db.RetrieveUserCookieById(db, userModel.Name)
	err = user.CheckCookieEaGateAccess(client, cookie)
	if err != nil {
		return
	}

	user.AddCookiesToJar(client.Client.Jar, []*http.Cookie{cookie})
	return
}