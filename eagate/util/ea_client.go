package util

import (
	"bytes"
	"context"
	"github.com/golang/glog"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

type EaClient struct {
	Client *http.Client
	username string
	ActiveCookie string
}

var (
	s *semaphore.Weighted
)

// GenerateClient will generate a http.client that is
// used by this library.
func GenerateClient() EaClient {
	glog.Infoln("generating new eaclient")
	jar := NewJar()

	if s == nil {
		s = semaphore.NewWeighted(1024)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
		Transport: ClientRateLimiter {
			http.DefaultTransport,
			s,
		},
	}
	return EaClient{client, "", ""}
}

type ClientRateLimiter struct {
	Proxy http.RoundTripper
	WeightedSemaphore *semaphore.Weighted
}

func (crl ClientRateLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	crl.WeightedSemaphore.Acquire(context.Background(), 1)
	defer crl.WeightedSemaphore.Release(1)
	return crl.Proxy.RoundTrip(req)
}

func (client *EaClient) SetUsername(un string) {
	client.username = strings.ToLower(un)
	glog.Infof("client username changed to %s\n", client.username)
}

func (client *EaClient) GetUsername() string {
	return client.username
}

func (client *EaClient) SetEaCookie(cookie *http.Cookie) {
	eagate, _ := url.Parse("https://p.eagate.573.jp")
	var cookies []*http.Cookie
	cookie.Domain = "p.eagate.573.jp"
	cookies = append(cookies, cookie)

	client.Client.Jar.SetCookies(eagate, cookies)
	client.ActiveCookie = cookie.String()
	glog.Infof("eacookie changed for username %s\n", client.username)
}


func (client *EaClient) GetEaCookie() *http.Cookie {
	eagate, _ := url.Parse("https://p.eagate.573.jp")
	currCookie := client.Client.Jar.Cookies(eagate)
	if len(currCookie) == 0 {
		return nil
	}
	return currCookie[0]
}

func (client *EaClient) LoginState() bool {
	res, err := client.Client.Get("https://p.eagate.573.jp/gate/p/mypage/index.html")
	if err != nil || res.StatusCode != 200 {
		glog.Warningf("loginstate for %s is false, status %d\n", client.username, res.StatusCode)
		return false
	}

	currCookie := client.GetEaCookie()
	if currCookie != nil && currCookie.String() != client.ActiveCookie {
		glog.Infof("cookie for user %s changed\n", client.username)
		client.SetEaCookie(currCookie)
	}
	glog.Infof("cookie set for user %s\n", client.username)
	return true
}


// setClient is in place to replace the internal client with a test client
func (client *EaClient) SetTestClient(testServer *httptest.Server, responseMap map[string]string) {
	client.Client = testServer.Client()

	client.Client.Transport = TestClientProxy{
		http.DefaultTransport,
		responseMap,
	}
}

type TestClientProxy struct {
	Proxy http.RoundTripper
	ResponseMap map[string]string
}

func (tcp TestClientProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if file, ok := tcp.ResponseMap[req.URL.String()]; ok {
		fileContents, err := ioutil.ReadFile(file)
		if err == nil {
			r := &http.Response{
				Status:           "OK",
				StatusCode:       http.StatusOK,
				Body:             ioutil.NopCloser(bytes.NewReader(fileContents)),
				ContentLength:    int64(len(fileContents)),
			}
			return r, nil
		}
	}
	r := &http.Response{
		Status:           "BadRequest",
		StatusCode:       http.StatusBadRequest,
		Body:             ioutil.NopCloser(strings.NewReader("bad")),
		ContentLength:    int64(len("bad")),
	}
	return r, nil
}