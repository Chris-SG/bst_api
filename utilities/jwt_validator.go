package utilities

import (
	"encoding/json"
	"errors"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/glog"
	"net/http"
	"strings"
)

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N string `json:"n"`
	E string `json:"e"`
	X5c []string `json:"x5c"`
}

func GetJWTMiddleware() *jwtmiddleware.JWTMiddleware{

	return jwtmiddleware.New(jwtmiddleware.Options {
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			aud := authClientAudience
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New("Invalid audience.")
			}
			// Verify 'iss' claim
			iss := authClientIssuer
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
		Extractor: jwtmiddleware.FromAuthHeader,
	})
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(authClientIssuer + ".well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return cert, err
	}

	return cert, nil
}

// profileFromToken will extract the user profile from the
// request JWT token. This contains data used to validate
// the user against an eagate account.
func ProfileFromToken(r *http.Request) map[string]interface{} {
	token, err := jwtmiddleware.FromAuthHeader(r)
	if err != nil {
		panic(err)
	}
	splitToken := strings.Split(token, ".")

	decodedToken, err := jwt.DecodeSegment(splitToken[1])
	if err != nil {
		panic(err)
	}

	tokenMap := make(map[string]interface{})

	err = json.Unmarshal(decodedToken, &tokenMap)
	if err != nil {
		panic(err)
	}

	if impersonateUser := r.Header.Get("Impersonate-User"); len(impersonateUser) > 0 {
		val, ok := tokenMap["sub"].(string)
		if ok {
			val = strings.ToLower(val)
			if UserHasScopes(val, []string{"impersonate"}) {
				glog.Infof("%s is impersonating %s", val, impersonateUser)
				tokenMap["sub"] = impersonateUser
			}
		}
	}

	return tokenMap
}