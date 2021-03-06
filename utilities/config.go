package utilities

import (
	"flag"
	"github.com/chris-sg/bst_api/db"
	"github.com/golang/glog"
)

var (
	authClientIssuer string
	authClientAudience string

	ServeHost string
	ServePort string
	ApiBase string

	DbMigration bool

	a0MgmtAudience string
	a0MgmtClientId string
	a0MgmtClientSecret string
)

func LoadConfig() {
	flag.StringVar(&authClientIssuer, "issuer", "", "the issuer for auth server.")
	flag.StringVar(&authClientAudience, "audience", "", "the audience for auth server.")

	flag.StringVar(&a0MgmtAudience, "a0mgmtaudience", "", "audience for auth0 management.")
	flag.StringVar(&a0MgmtClientId, "a0mgmtclientid", "", "client id for auth0 management.")
	flag.StringVar(&a0MgmtClientSecret, "a0mgmtclientsecret", "", "client secret for auth0 management.")

	flag.StringVar(&ServeHost, "host", "", "the host.")
	flag.StringVar(&ServePort, "port", "8443", "the port.")
	flag.StringVar(&ApiBase, "apibase", "/", "bst api base path.")

	flag.BoolVar(&DbMigration, "dbmigrate", false, "run db migration and exit.")

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

	err := db.OpenDb("postgres", user, password, dbname, host, maxIdleConnections)
	if err != nil {
		glog.Fatalln("Failed to open db!")
		panic(err)
	}

}