package dataservice

import (
	"database/sql"
	"dddfile/util"
	"fmt"

	_ "github.com/lib/pq"
)

var _db *sql.DB

func Init() {

	var (
		host     = util.GetEnvVar("DATABASE_HOST")
		port     = util.GetEnvVar("DATABASE_PORT")
		username = util.GetEnvVar("DATABASE_USERNAME")
		password = util.GetEnvVar("DATABASE_PASSWORD")
		database = util.GetEnvVar("DATABASE_NAME")
	)

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, port, username, password, database)
	if util.GetEnvVar("GIN_MODE") != "release" {
		psqlconn = psqlconn + " sslmode=disable"
	}

	// open database
	db, err := sql.Open("postgres", psqlconn)
	util.CheckError(err)

	// // close database
	// defer db.Close()

	_db = db
}

func GetDb() *sql.DB {
	return _db
}
