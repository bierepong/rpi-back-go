package main

import (
	"database/sql"
	"fmt"
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

		_, errExec := db.Exec("create table game_data (username text not null primary key, score integer not null);")
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

func (db *GameDB) insertUser(username string, score int) (string, int, error) {
	stmt, errPrepare := db.Prepare("insert into game_data(username, score) values(?, ?)")
	if errPrepare != nil {
		return "", 0, errPrepare
	}
	defer closeStatement(stmt)

	_, errExec := stmt.Exec(username, score)
	if errExec != nil {
		return "", 0, errExec
	}

	return username, score, nil
}

func (db *GameDB) userExists(username string) (bool, error) {
	stmt, errPrepare := db.Prepare("select count(username) from game_data where username=?")
	if errPrepare != nil {
		return false, errPrepare
	}
	defer closeStatement(stmt)

	rows, errQuery := stmt.Query(username)
	if errQuery != nil {
		return false, errQuery
	}
	defer closeRows(rows)

	for rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			return false, err
		}

		switch {
		case count == 0:
			return false, nil
		case count == 1:
			return true, nil
		case count > 1:
			return false, fmt.Errorf("more than one result found")
		default:
			return false, fmt.Errorf("unknown error")
		}
	}

	return false, fmt.Errorf("unknown error")
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

func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		log.WithError(err).Error("error when closing rows")
	}
}
