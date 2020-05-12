package db

import (
	"fmt"
	"github.com/chris-sg/bst_api/db/api_db"
	"github.com/chris-sg/bst_api/db/db_builder"
	"github.com/chris-sg/bst_api/db/ddr_db"
	"github.com/chris-sg/bst_api/db/drs_db"
	"github.com/chris-sg/bst_api/db/user_db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	db *gorm.DB
	apiDbComm api_db.ApiDbCommunication
	ddrDbComm ddr_db.DdrDbCommunication
	drsDbComm drs_db.DrsDbCommunication
	idleConnectionLimit int
	migrator db_builder.DbMigrator
	userDbComm user_db.UserDbCommunication
)

func OpenDb(dialect string, user string, password string, dbname string, host string, maxIdleConnections int) (err error) {
	availableDialects := []string{"postgres"}

	found := false
	for _, d := range availableDialects {
		if d == dialect {
			found = true
			break
		}
	}
	if !found {
		err = fmt.Errorf("dialect %s not supported", dialect)
		return
	}

	if dialect == "postgres" {
		openDbPostgres(user, password, dbname, host, maxIdleConnections)
	}

	return nil
}

func openDbPostgres(user string, password string, dbname string, host string, maxIdleConnections int) (err error) {
	connStr := fmt.Sprintf( "user=%s password=%s dbname=%s host=%s sslmode=disable", user, password, dbname, host)

	newDb, err := gorm.Open("postgres", connStr)
	if err != nil {
		return
	}
	db = newDb
	db.DB().SetMaxIdleConns(maxIdleConnections)
	idleConnectionLimit = maxIdleConnections

	apiDbComm = api_db.CreateApiDbCommunicationPostgres(db)
	ddrDbComm = ddr_db.CreateDdrDbCommunicationPostgres(db)
	drsDbComm = drs_db.CreateDrsDbCommunicationPostgres(db)
	migrator = db_builder.CreateDbMigratorPostgres(db)
	userDbComm = user_db.CreateUserDbCommunicationPostgres(db)

	return
}

func GetDb() (*gorm.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("db connection has not been created, please use OpenDb()")
	}
	return db, nil
}

func GetApiDb() api_db.ApiDbCommunication {
	return apiDbComm
}

func GetDdrDb() ddr_db.DdrDbCommunication {
	return ddrDbComm
}

func GetDrsDb() drs_db.DrsDbCommunication {
	return drsDbComm
}

func GetMigrator() db_builder.DbMigrator {
	return migrator
}

func GetUserDb() user_db.UserDbCommunication {
	return userDbComm
}

func GetIdleConnectionLimit() int {
	return idleConnectionLimit
}
