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

	dbPath := os.Getenv("BEERPONG_DB_PATH")
	// Get default value
	if dbPath == "" {
		dbPath = "./test_db.sqlite"
	}

	log.WithField("dbPath", dbPath).Info("database selected")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		vanillaDB, errOpen := sql.Open("sqlite3", dbPath)
		if errOpen != nil {
			log.WithError(errOpen).Fatal("error when opening database")
		}
		db = &GameDB{vanillaDB}

		_, errExec := db.Exec("create table game_data (username text not null primary key);")
		if errExec != nil {
			log.WithError(errExec).Fatal("error on create table init")
		}
	} else {
		vanillaDB, errOpen := sql.Open("sqlite3", dbPath)
		if errOpen != nil {
			log.WithError(errOpen).Fatal("error when opening database")
		}
		db = &GameDB{vanillaDB}
	}

	dbInstance = db
	return dbInstance
}

func (db *GameDB) insertUser(username string) (string, error) {
	stmt, errPrepare := db.Prepare("insert into game_data(username) values(?)")
	if errPrepare != nil {
		return "", errPrepare
	}
	defer closeStatement(stmt)

	_, errExec := stmt.Exec(username)
	if errExec != nil {
		return "", errExec
	}

	return username, nil
}

func (db *GameDB) close() {
	if err := db.Close(); err != nil {
		log.WithError(err).Error("error when closing database")
	}
}

func closeStatement(stmt *sql.Stmt) {
	if err := stmt.Close(); err != nil {
		log.WithError(err).Error("error when closing statement")
	}
}
