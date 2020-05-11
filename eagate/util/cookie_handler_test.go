package util

import "testing"

func TestCookie(t *testing.T) {
	client := GenerateClient()
	cookieText := "M573SSID=ab4d4e5a-38a3-4f23-aa9f-90cbe40419c1; Path=/; Domain=p.eagate.573.jp; Expires=Tue, 24 Mar 2020 00:35:26 GMT; HttpOnly; Secure"
	cookie := CookieFromRawCookie(cookieText)
	if cookieText != cookie.Raw {
		t.Errorf("cookieText does not match cookie.Raw: %s and %s", cookieText, cookie.Raw)
	}
	if cookieText != cookie.String() {
		t.Errorf("cookieText does not match cookie.String(): %s and %s", cookieText, cookie.String())
	}

	client.SetEaCookie(cookie)
	setCookie := client.GetEaCookie()
	if cookieText != setCookie.Raw {
		t.Errorf("cookieText does not match setCookie.Raw: %s and %s", cookieText, setCookie.Raw)
	}
	if cookieText != setCookie.String() {
		t.Errorf("cookieText does not match setCookie.String(): %s and %s", cookieText, setCookie.String())
	}
}

