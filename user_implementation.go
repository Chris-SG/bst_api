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