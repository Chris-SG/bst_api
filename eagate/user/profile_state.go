package user

import (
	"github.com/chris-sg/bst_api/db"
	"github.com/chris-sg/bst_api/eagate/util"
	"github.com/chris-sg/bst_api/models/user_models"
	"github.com/chris-sg/bst_api/utilities"
	bst_models "github.com/chris-sg/bst_server_models"
	"github.com/golang/glog"
)

// createClientForUser will generate a http client for the provided user
// model. This is intended to only be used for this specific user model,
// as it will use cookies from the database for eagate integration.
func CreateClientForUser(userModel user_models.User) (client util.EaClient, err bst_models.Error) {
	err = bst_models.ErrorOK
	client = util.GenerateClient()
	client.SetUsername(userModel.Name)
	cookie, errs := db.GetUserDb().RetrieveUserCookieStringByUserId(userModel.Name)
	if utilities.PrintErrors("failed to retrieve cookie:", errs) {
		err = bst_models.ErrorNoCookie
		return
	}
	if len(cookie) == 0 {
		err = bst_models.ErrorNoCookie
		return
	}
	client.SetEaCookie(util.CookieFromRawCookie(cookie))
	if !client.LoginState() {
		err = bst_models.ErrorBadCookie
	}
	return
}

func ProfileEaSubscriptionState(client util.EaClient) (subscriptionType string, err bst_models.Error) {
	err = bst_models.ErrorOK
	const paybookResource = "/payment/mybook/paybook.html"
	PaybookUri := util.BuildEaURI(paybookResource)

	document, err := util.GetPageContentAsGoQuery(client.Client, PaybookUri)
	eaSubSelection := document.Find("div#id_paybook_all .cl_course_name").First()

	if eaSubSelection == nil {
		err = bst_models.ErrorBadCookie
		return
	}

	subscriptionType = eaSubSelection.Text()
	return
}

func RunUpdatesOnAllEaUsers() {
	users, errs := db.GetUserDb().RetrieveUsersForUpdate()
	if len(errs) > 0 {
		glog.Errorf("failed to update users due to selection error")
		glog.Error(errs)
		return
	}

	for _, user := range users {
		func() {
			client, err := CreateClientForUser(user)
			defer client.UpdateCookie()
			if err.Equals(bst_models.ErrorBadCookie) || err.Equals(bst_models.ErrorNoCookie) {
				user.Cookie = ""
				user.Expiration = 0
				errs = db.GetUserDb().UpdateUser(user)
				if len(errs) > 0 {
					glog.Errorf("failed to save user %s", user.WebUser)
					glog.Error(errs)
				}
				return
			}
			if !err.Equals(bst_models.ErrorOK) {
				glog.Warningf("client upgrade for %s failed: %s", user, err.Message)
				return
			}
			if !client.LoginState() {
				user.Cookie = ""
				user.Expiration = 0
				errs = db.GetUserDb().UpdateUser(user)
				if len(errs) > 0 {
					glog.Errorf("failed to save user %s", user.WebUser)
					glog.Error(errs)
				}
				return
			}
			user.Cookie = client.GetEaCookie().String()
			user.Expiration = client.GetEaCookie().Expires.UnixNano() / 1000
			subState, err := ProfileEaSubscriptionState(client)
			if !err.Equals(bst_models.ErrorOK) {
				glog.Warningf("client upgrade for %s failed: %s", user, err.Message)
				return
			}
			user.EaSubscription = subState
			errs = db.GetUserDb().UpdateUser(user)
			if len(errs) > 0 {
				glog.Errorf("failed to save user %s", user.WebUser)
				glog.Error(errs)
			}
		}()
	}
}