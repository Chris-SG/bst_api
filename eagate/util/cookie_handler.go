package util

import (
	"github.com/golang/glog"
	"golang.org/x/net/http/httpguts"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CookieFromRawCookie(rawCookie string) *http.Cookie {
	return parseRawCookie(rawCookie)
}

func parseRawCookie(rawCookie string) *http.Cookie {
	parts := strings.Split(strings.TrimSpace(rawCookie), ";")
	if len(parts) == 1 && parts[0] == "" {
		glog.Errorln("attempted to parse empty rawCookie")
		return nil
	}
	glog.Infof("parsing raw cookie %s", rawCookie[0:9])
	parts[0] = strings.TrimSpace(parts[0])
	j := strings.Index(parts[0], "=")
	if j < 0 {
		glog.Errorf("raw cookie %s does not contain '='\n", rawCookie[0:9])
		return nil
	}
	name, value := parts[0][:j], parts[0][j+1:]
	if !isCookieNameValid(name) {
		return nil
	}
	value, ok := parseCookieValue(value, true)
	if !ok {
		glog.Errorf("failed to parsecookievalue on %s\n", rawCookie[0:9])
		return nil
	}
	c := &http.Cookie{
		Name:  name,
		Value: value,
		Raw:   rawCookie,
	}
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.TrimSpace(parts[i])
		if len(parts[i]) == 0 {
			continue
		}

		attr, val := parts[i], ""
		if j := strings.Index(attr, "="); j >= 0 {
			attr, val = attr[:j], attr[j+1:]
		}
		lowerAttr := strings.ToLower(attr)
		val, ok = parseCookieValue(val, false)
		if !ok {
			c.Unparsed = append(c.Unparsed, parts[i])
			continue
		}
		switch lowerAttr {
		case "samesite":
			lowerVal := strings.ToLower(val)
			switch lowerVal {
			case "lax":
				c.SameSite = http.SameSiteLaxMode
			case "strict":
				c.SameSite = http.SameSiteStrictMode
			case "none":
				c.SameSite = http.SameSiteNoneMode
			default:
				c.SameSite = http.SameSiteDefaultMode
			}
			continue
		case "secure":
			c.Secure = true
			continue
		case "httponly":
			c.HttpOnly = true
			continue
		case "domain":
			c.Domain = val
			continue
		case "max-age":
			secs, err := strconv.Atoi(val)
			if err != nil || secs != 0 && val[0] == '0' {
				break
			}
			if secs <= 0 {
				secs = -1
			}
			c.MaxAge = secs
			continue
		case "expires":
			c.RawExpires = val
			exptime, err := time.Parse(time.RFC1123, val)
			if err != nil {
				exptime, err = time.Parse("Mon, 02-Jan-2006 15:04:05 MST", val)
				if err != nil {
					glog.Warningf("parsing time on %s failed: %s\n", rawCookie[0:9], val)
					c.Expires = time.Time{}
					break
				}
			}
			c.Expires = exptime.UTC()
			continue
		case "path":
			c.Path = val
			continue
		}
		c.Unparsed = append(c.Unparsed, parts[i])
	}

	return c
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}