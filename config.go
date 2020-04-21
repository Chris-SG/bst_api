package main

import (
	"flag"
	"github.com/chris-sg/eagate_db"
	"github.com/golang/glog"
)

var (
	authClientIssuer string
	authClientAudience string

	serveHost string
	servePort string
	apiBase string

	dbMigration bool
)

func LoadConfig() {
	flag.StringVar(&authClientIssuer, "issuer", "", "the issuer for auth server.")
	flag.StringVar(&authClientAudience, "audience", "", "the audience for auth server.")

	flag.StringVar(&serveHost, "host", "", "the host.")
	flag.StringVar(&servePort, "port", "8443", "the port.")
	flag.StringVar(&apiBase, "apibase", "/", "bst api base path.")

	flag.BoolVar(&dbMigration, "dbmigrate", false, "run db migration and exit.")

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

	glog.Infoln("Done!")

	err := eagate_db.OpenDb("postgres", user, password, dbname, host, maxIdleConnections)
	if err != nil {
		glog.Fatalln("Failed to open db!")
		panic(err)
	}

}