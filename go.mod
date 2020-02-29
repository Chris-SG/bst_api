module github.com/chris-sg/bst_api

go 1.13

require (
	github.com/auth0/go-jwt-middleware v0.0.0-20190805220309-36081240882b
	github.com/chris-sg/bst_server_models v0.0.0-20200228075122-635b259117ff
	github.com/chris-sg/eagate v0.0.0
	github.com/chris-sg/eagate_db v0.0.0
	github.com/chris-sg/eagate_models v0.0.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
)

replace github.com/chris-sg/eagate v0.0.0 => ../eagate

replace github.com/chris-sg/eagate_db v0.0.0 => ../eagate_db

replace github.com/chris-sg/eagate_models v0.0.0 => ../eagate_models
