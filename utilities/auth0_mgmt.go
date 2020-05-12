package utilities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	token string
	expiration time.Time
)

func UserHasScopes(user string, scopes []string) bool {
	userScopes := GetUserScopes(user)

	correctScopes := 0
	for _, requiredScope := range scopes {
		for _, scope := range userScopes {
			if requiredScope == scope {
				correctScopes++
				break
			}
		}
	}

	return correctScopes == len(scopes)
}

func GetUserScopes(user string) (scopes []string) {
	type Source struct {
		SourceId string `json:"source_id"`
		SourceName string `json:"source_name"`
		SourceType string `json:"source_type"`
	}
	type Permission struct {
		PermissionName string `json:"permission_name"`
		Description string `json:"description"`
		ResourceServerName string `json:"resource_server_name"`
		ResourceServerIdentifier string `json:"resource_server_identifier"`
		Sources []Source `json:"sources"`
	}

	if !validateToken() {
		return
	}
	uri, _ := url.Parse(authClientIssuer)
	uri.Path = fmt.Sprintf("/api/v2/users/%s/permissions", user)

	req := &http.Request{
		Method:           http.MethodGet,
		URL:              uri,
		Header:			  make(map[string][]string),
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	client := http.Client{}
	r, err := client.Do(req)

	if err != nil {
		glog.Errorf("error retrieving scopes: %s", err.Error())
		return
	}

	if r.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			glog.Errorf("could not get scope for %s: %s", user, err.Error())
			return
		}
		glog.Errorf("could not get scope for user %s with error code %d: %s", user, r.StatusCode, string(bodyBytes))
		return
	}

	permissions := make([]Permission, 0)

	j := json.NewDecoder(r.Body)
	err = j.Decode(&permissions)

	if err != nil {
		glog.Errorf("error retrieving scopes: %s", err.Error())
		return
	}

	for _, permission := range permissions {
		scopes = append(scopes, permission.PermissionName)
	}
	return
}

func validateToken() bool {
	splitToken := strings.Split(token, ".")
	if len(splitToken) == 0 || time.Now().UnixNano() > expiration.UnixNano() {
		if !refreshToken() {
			return false
		}
	}

	return true
}

func refreshToken() bool {
	uri, _ := url.Parse(authClientIssuer + "oauth/token")

	req := &http.Request{
		Method:           http.MethodPost,
		URL:              uri,
		Header:			  make(map[string][]string),
	}
	req.Header.Add("content-type", "application/json")

	type TokenReq struct {
		ClientId string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Audience string `json:"audience"`
		GrantType string `json:"grant_type"`
	}
	type TokenRes struct {
		AccessToken string `json:"access_token"`
		Scope string `json:"scope"`
		ExpiresIn time.Duration `json:"expires_in"`
		TokenType string `json:"token_type"`
	}

	tokenRequest := TokenReq{
		ClientId:     a0MgmtClientId,
		ClientSecret: a0MgmtClientSecret,
		Audience:     a0MgmtAudience,
		GrantType:    "client_credentials",
	}

	body, _ := json.Marshal(tokenRequest)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	client := http.Client{}
	r, err := client.Do(req)

	if err != nil {
		glog.Errorf("error creating mgmt token: %s", err.Error())
		return false
	}

	if r.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			glog.Errorf("could not read mgmt token body: %s", err.Error())
			return false
		}
		glog.Errorf("mgmt token request responded with status code %d: %s", r.StatusCode, string(bodyBytes))
		return false
	}

	response := TokenRes{}
	d := json.NewDecoder(r.Body)
	err = d.Decode(&response)
	if err != nil {
		glog.Errorf("failed to decode response: %s", err.Error())
		return false
	}

	token = response.AccessToken
	expiration = time.Now()
	expiration.Add(response.ExpiresIn * time.Second)

	return true
}
