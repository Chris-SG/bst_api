package user_db

import (
	"fmt"
	"github.com/chris-sg/bst_api/models/user_models"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"net/http"
	"strings"
	"time"
)

type UserDbCommunication interface {
	SetCookieForUser(userId string, cookie *http.Cookie) (errs []error)
	SetSubscriptionForUser(userId string, sub string) (errs []error)
	RetrieveUserByUserId(userId string) (user user_models.User, errs []error)
	RetrieveUserByWebId(webUserId string) (user user_models.User, errs []error)
	RetrieveUserCookieStringByUserId(userId string) (cookie string, errs []error)
	SetWebUserForEaUser(userId string, webUserId string) (errs []error)
	UpdateUser(user user_models.User) (errs []error)

	RetrieveUsersForUpdate() (users []user_models.User, errs []error)
	RetrieveRandomHelper() (user user_models.User, errs []error)
}

func CreateUserDbCommunicationPostgres(db *gorm.DB) UserDbCommunicationPostgres {
	return UserDbCommunicationPostgres{db}
}

type UserDbCommunicationPostgres struct {
	db *gorm.DB
}

// TODO: use cookie string instead
func (dbcomm UserDbCommunicationPostgres) SetCookieForUser(userId string, cookie *http.Cookie) (errs []error) {
	glog.Infof("SetCookieForUser for user id %s\n", userId)
	userId = strings.ToLower(userId)
	eaGateUser, errs := dbcomm.RetrieveUserByUserId(userId)
	if len(errs) > 0 {
		return
	}
	if eaGateUser.Name == "" {
		eaGateUser = user_models.User{}
	}
	eaGateUser.Name = strings.ToLower(eaGateUser.Name)
	eaGateUser.Cookie = cookie.String()
	eaGateUser.Expiration = cookie.Expires.UnixNano() / 1000

	resultDb := dbcomm.db.Save(eaGateUser)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm UserDbCommunicationPostgres) SetSubscriptionForUser(userId string, sub string) (errs []error) {
	glog.Infof("SetCookieForUser for user id %s\n", userId)
	userId = strings.ToLower(userId)
	eaGateUser, errs := dbcomm.RetrieveUserByUserId(userId)
	if len(errs) > 0 {
		return
	}
	if eaGateUser.Name == "" {
		eaGateUser = user_models.User{}
	}

	eaGateUser.Name = strings.ToLower(eaGateUser.Name)
	eaGateUser.EaSubscription = sub

	resultDb := dbcomm.db.Save(eaGateUser)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm UserDbCommunicationPostgres) RetrieveUserByUserId(userId string) (user user_models.User, errs []error) {
	glog.Infof("RetrieveUserByUserId for user id %s\n", userId)
	userId = strings.ToLower(userId)
	resultDb := dbcomm.db.Model(&user_models.User{}).Where("account_name = ?", userId).First(&user)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm UserDbCommunicationPostgres) RetrieveUserByWebId(webUserId string) (user user_models.User, errs []error) {
	glog.Infof("RetrieveUserByWebId for web id %s\n", webUserId)
	webUserId = strings.ToLower(webUserId)
	resultDb := dbcomm.db.Model(&user_models.User{}).Where("web_user = ?", webUserId).First(&user)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm UserDbCommunicationPostgres) RetrieveUserCookieStringByUserId(userId string) (cookie string, errs []error) {
	glog.Infof("RetrieveUserCookieStringByUserId for user id %s\n", userId)
	userId = strings.ToLower(userId)
	eaGateUser, errs := dbcomm.RetrieveUserByUserId(userId)
	if len(errs) > 0 {
		return
	}

	timeNow := time.Now().UnixNano() / 1000
	if len(eaGateUser.Cookie) == 0 || eaGateUser.Expiration < timeNow {
		errs = append(errs, fmt.Errorf("cookie for user id %s was not found or expired", userId))
		return
	}
	cookie = eaGateUser.Cookie
	glog.Infof("RetrieveUserCookieById: retrieved cookie for user id %s\n", userId)
	return
}

func (dbcomm UserDbCommunicationPostgres) SetWebUserForEaUser(userId string, webUserId string) (errs []error) {
	glog.Infof("SetWebUserForUser: user id %s, web id %s\n", userId, webUserId)
	userId = strings.ToLower(userId)
	webUserId = strings.ToLower(webUserId)
	eaGateUser, errs := dbcomm.RetrieveUserByUserId(userId)
	if len(errs) > 0 {
		return
	}

	eaGateUser.WebUser = webUserId
	resultDb := dbcomm.db.Save(eaGateUser)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm UserDbCommunicationPostgres) RetrieveUsersForUpdate() (users []user_models.User, errs []error) {
	resultDb := dbcomm.db.Model(&user_models.User{}).Where("login_cookie <> ?", "").Scan(&users)
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm UserDbCommunicationPostgres) UpdateUser(user user_models.User) (errs []error) {
	resultDb := dbcomm.db.Save(&user)
	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}

func (dbcomm UserDbCommunicationPostgres) RetrieveRandomHelper() (user user_models.User, errs []error) {
	resultDb := dbcomm.db.Model(&user_models.User{}).
		Where("login_cookie <> ?", "").
		Where("subscription in (?)", []string{"e-amusement ベーシックコース"}).
		Order("rand()").
		First(&user)

	errors := resultDb.GetErrors()
	if errors != nil && len(errors) != 0 {
		errs = append(errs, errors...)
	}
	return
}