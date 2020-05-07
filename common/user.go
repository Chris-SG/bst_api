package common

import (
	"encoding/json"
	"fmt"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/chris-sg/eagate/user"
	"github.com/chris-sg/eagate/util"
	"github.com/chris-sg/eagate_db"
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
	users, errMsg, err := TryGetEagateUsers(r)
	if err != nil {
		status := utilities.WriteStatus("bad", errMsg)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(bytes)
		return
	}

	eagateUsers := make([]bst_models.EagateUser, 0)

	for i, _ := range users {
		eagateUser := bst_models.EagateUser{
			Username: users[i].Name,
			Expired:  users[i].Expiration < time.Now().UnixNano()/1000,
		}
		if !eagateUser.Expired {
			client, _, err := CreateClientForUser(users[i])
			if err != nil || !client.LoginState() {
				fmt.Println(err)
				eagateUser.Expired = true
			}
		}
		eagateUsers = append(eagateUsers, eagateUser)
	}

	bytes, err := json.Marshal(eagateUsers)
	if err != nil {
		status := utilities.WriteStatus("bad", "marshal_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(bytes)
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
		status := utilities.WriteStatus("bad", "jwt_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}
	val = strings.ToLower(val)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("%s\n", err.Error())
		status := utilities.WriteStatus("bad", "bad_api_req")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	loginRequest := bst_models.LoginRequest{}
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		glog.Warningf("failed to decode login request for %s: %s\n", loginRequest.Username, err.Error())
		status := utilities.WriteStatus("bad", "unmarshal_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	glog.Infof("user %s attempting to login to eagate\n", loginRequest.Username)
	client := util.GenerateClient()

	cookie, err := user.GetCookieFromEaGate(loginRequest.Username, loginRequest.Password, loginRequest.OneTimePassword, client)
	if err != nil {
		status := utilities.WriteStatus("bad", "bad_cookie")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	errs := eagate_db.GetUserDb().SetCookieForUser(loginRequest.Username, cookie)
	if len(errs) > 0 {
		status := utilities.WriteErrorStatus(errs)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}
	errs = eagate_db.GetUserDb().SetWebUserForEaUser(loginRequest.Username, val)
	if len(errs) > 0 {
		status := utilities.WriteErrorStatus(errs)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	status := utilities.WriteStatus("ok", "loaded cookie for user")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusOK)
	rw.Write(bytes)
	return
}

func LogoutPost(rw http.ResponseWriter, r *http.Request) {
	tokenMap := utilities.ProfileFromToken(r)

	val, ok := tokenMap["sub"].(string)
	if !ok {
		status := utilities.WriteStatus("bad", "jwt_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}
	val = strings.ToLower(val)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		status := utilities.WriteStatus("bad", "bad_api_req")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	logoutRequest := bst_models.LogoutRequest{}
	err = json.Unmarshal(body, &logoutRequest)
	if err != nil {
		status := utilities.WriteStatus("bad", "unmarshal_err")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	user, errs := eagate_db.GetUserDb().RetrieveUserByUserId(logoutRequest.Username)
	if len(errs) > 0 {
		status := utilities.WriteErrorStatus(errs)
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(bytes)
		return
	}

	if user.WebUser == val {
		errs := eagate_db.GetUserDb().SetWebUserForEaUser(user.Name, "")
		if len(errs) > 0 {
			status := utilities.WriteErrorStatus(errs)
			bytes, _ := json.Marshal(status)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(bytes)
			return
		}

		status := utilities.WriteStatus("ok", "user unlinked")
		bytes, _ := json.Marshal(status)
		rw.WriteHeader(http.StatusOK)
		rw.Write(bytes)
		return
	}

	status := utilities.WriteStatus("bad", "no_user")
	bytes, _ := json.Marshal(status)
	rw.WriteHeader(http.StatusUnauthorized)
	rw.Write(bytes)
	return
}