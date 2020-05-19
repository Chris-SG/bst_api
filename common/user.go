package common

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/user"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// LoginGet will retrieve any relations between the requester and the
// database. This may produce multiple relations in the case a user
// has linked multiple accounts. Any stored cookies will be nullified.
func LoginGet(rw http.ResponseWriter, r *http.Request) {
	users, err := TryGetEagateUsers(r)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	eagateUsers := make([]bst_models.EagateUser, 0)

	for i := range users {
		eagateUser := bst_models.EagateUser{
			Username: users[i].Name,
			Expired:  users[i].Expiration < time.Now().UnixNano()/1000,
		}
		if !eagateUser.Expired {
			func() {
				client, err := user.CreateClientForUser(users[i])
				defer client.UpdateCookie()
				if !err.Equals(bst_models.ErrorOK) || !client.LoginState() {
					fmt.Println(err)
					eagateUser.Expired = true
				}
			}()
		}
		eagateUsers = append(eagateUsers, eagateUser)
	}

	bytes, e := json.Marshal(eagateUsers)
	if e != nil {
		glog.Warning(e)
		utilities.RespondWithError(rw, bst_models.ErrorJsonEncode)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
}

// LoginPost will attempt to login to eagate using the supplied credentials.
// This will not check the current user state, which if exists, will have
// its cookie updated and web user replaced (if applicable).
func LoginPost(rw http.ResponseWriter, r *http.Request) {
	tokenMap := utilities.ProfileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		utilities.RespondWithError(rw, bst_models.ErrorJwtProfile)
		return
	}
	val = strings.ToLower(val)

	body, e := ioutil.ReadAll(r.Body)
	if e != nil {
		glog.Errorf("%s\n", e.Error())
		utilities.RespondWithError(rw, bst_models.ErrorBadBody)
		return
	}

	loginRequest := bst_models.LoginRequest{}
	e = json.Unmarshal(body, &loginRequest)
	if e != nil {
		glog.Warningf("failed to decode login request for %s: %s\n", loginRequest.Username, e.Error())
		utilities.RespondWithError(rw, bst_models.ErrorJsonDecode)
		return
	}

	glog.Infof("user %s attempting to login to eagate\n", loginRequest.Username)
	client := util.GenerateClient()
	defer client.UpdateCookie()

	err := user.GetCookieFromEaGate(loginRequest.Username, loginRequest.Password, loginRequest.OneTimePassword, client)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	sub, err := user.ProfileEaSubscriptionState(client)
	if !err.Equals(bst_models.ErrorOK) {
		utilities.RespondWithError(rw, err)
		return
	}

	errs := db.GetUserDb().SetSubscriptionForUser(loginRequest.Username, sub)
	if utilities.PrintErrors("could not set ea subscription for user", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorWriteWebUser)
		return
	}
	errs = db.GetUserDb().SetWebUserForEaUser(loginRequest.Username, val)
	if utilities.PrintErrors("could not set web user for ea user", errs) {
		utilities.RespondWithError(rw, bst_models.ErrorWriteWebUser)
		return
	}

	utilities.RespondWithError(rw, bst_models.ErrorOK)
	return
}

func LogoutPost(rw http.ResponseWriter, r *http.Request) {
	tokenMap := utilities.ProfileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		utilities.RespondWithError(rw, bst_models.ErrorJwtProfile)
		return
	}
	val = strings.ToLower(val)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utilities.RespondWithError(rw, bst_models.ErrorBadBody)
		return
	}

	logoutRequest := bst_models.LogoutRequest{}
	err = json.Unmarshal(body, &logoutRequest)
	if err != nil {
		utilities.RespondWithError(rw, bst_models.ErrorJsonDecode)
		return
	}

	user, errs := db.GetUserDb().RetrieveUserByUserId(logoutRequest.Username)
	if len(errs) > 0 {
		utilities.RespondWithError(rw, bst_models.ErrorReadWebUser)
		return
	}

	if user.WebUser == val {
		errs := db.GetUserDb().SetWebUserForEaUser(user.Name, "")
		if len(errs) > 0 {
			utilities.RespondWithError(rw, bst_models.ErrorWriteWebUser)
			return
		}

		utilities.RespondWithError(rw, bst_models.ErrorOK)
		return
	}

	utilities.RespondWithError(rw, bst_models.ErrorNoEaUser)
	return
}

func ForceUpdatePost(rw http.ResponseWriter, r *http.Request) {
	requiredScopes := []string{"update:database"}
	tokenMap := utilities.ProfileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		utilities.RespondWithError(rw, bst_models.ErrorJwtProfile)
		return
	}
	val = strings.ToLower(val)
	if !utilities.UserHasScopes(val, requiredScopes) {
		glog.Warningf(
			"user %s tried to update users, but did not have required scopes %s",
			val,
			strings.Join(requiredScopes, ","))
		utilities.RespondWithError(rw, bst_models.ErrorScope)
		return
	}

	user.RunUpdatesOnAllEaUsers()

	utilities.RespondWithError(rw, bst_models.ErrorOK)
	return
}