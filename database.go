package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type GameDB struct {
	*sql.DB
}

var dbInstance *GameDB

func getDbClient() *GameDB {
	var db *GameDB

	if dbInstance != nil {
		return dbInstance
	}

	if _, err := os.Stat("./test_db.sqlite"); os.IsNotExist(err) {
		vanillaDB, errOpen := sql.Open("sqlite3", "./test_db.sqlite")
		if errOpen != nil {
			log.WithError(errOpen).Fatal("error when opening database")
		}
		db = &GameDB{vanillaDB}
		//defer db.Close()

		_, errExec := db.Exec("create table game_data (username text not null primary key);")
		if errExec != nil {
			log.WithError(errExec).Fatal("error on create table init")
		}
	} else {
		vanillaDB, errOpen := sql.Open("sqlite3", "./test_db.sqlite")
		if errOpen != nil {
			log.WithError(errOpen).Fatal("error when opening database")
		}
		db = &GameDB{vanillaDB}
		//defer db.Close()
	}

	dbInstance = db
	return dbInstance
}

func (db *GameDB) insertUser(username string) (string, error) {
	stmt, errPrepare := db.Prepare("insert into game_data(username) values(?)")
	if errPrepare != nil {
		return "", errPrepare
	}
	defer stmt.Close()

	_, errExec := stmt.Exec(username)
	if errExec != nil {
		return "", errExec
	}

	return username, nil
}
