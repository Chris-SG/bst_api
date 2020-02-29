package main

import (
	"flag"
	"github.com/chris-sg/eagate_db"
)

var (
	authClientIssuer string
	authClientAudience string

	serveHost string
	servePort string
	apiBase string
)

func LoadConfig() {
	flag.StringVar(&authClientIssuer, "issuer", "", "the issuer for auth server.")
	flag.StringVar(&authClientAudience, "audience", "", "the audience for auth server.")

	flag.StringVar(&serveHost, "host", "", "the host.")
	flag.StringVar(&servePort, "port", "8443", "the port.")
	flag.StringVar(&apiBase, "apibase", "/", "bst api base path.")

	var (
		user string
		password string
		dbname string
		host string
		maxIdleConnections int
	)

	flag.StringVar(&user, "dbuser", "", "the database user.")
	flag.StringVar(&password, "dbpass", "", "the database password.")
	flag.StringVar(&dbname, "dbname", "", "the database name.")
	flag.StringVar(&host, "dbhost", "", "the database host.")
	flag.IntVar(&maxIdleConnections, "dbmaxconns", 1, "the max idle db connections.")

	flag.Parse()

	_, err := eagate_db.OpenDb(user, password, dbname, host, maxIdleConnections)
	//db_builder.Create(db)
	if err != nil {
		panic(err)
	}

}