package main

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/json"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type (
	// A Database defines all the methods used to interact with the database.
	Database interface {
		// Close the database.
		Close() error
		// Exists checks if the user already exists.
		Exists(username string) (bool, error)
		// Insert a user in database.
		Insert(username string, score int) (string, int, error)
		// Update the given user's score.
		Update(username string, score int) (string, int, error)
	}

	// A User represents a database record.
	User struct {
		ID    string `json:"id"    storm:"id"`
		Name  string `json:"name"  storm:"unique"`
		Score int    `json:"score"`
	}

	database struct {
		db *storm.DB
	}
)

// Open returns a new database connection.
func Open(dbname string) (Database, error) {
	db, err := storm.Open(dbname, storm.Codec(json.Codec))
	if err != nil {
		return nil, errors.Wrap(err, "could not get database connection")
	}

	return &database{
		db: db,
	}, nil
}

func (o *database) Close() error {
	return o.db.Close()
}

func (o *database) Exists(username string) (bool, error) {
	n, err := o.db.Count(&User{Name: username}) // Name has uniqueness constraint.
	return n == 1, errors.Wrap(err, "could not check if username exists")
}

func (o *database) Insert(username string, score int) (string, int, error) {
	user := &User{
		ID:    uuid.Must(uuid.NewV4()).String(),
		Name:  username,
		Score: score,
	}
	return username, score, errors.Wrap(o.db.Save(user), "could not insert user")
}

func (o *database) Update(username string, score int) (string, int, error) {
	var user User
	if err := o.db.One("Name", username, &user); err != nil {
		return username, score, errors.Wrap(err, "update user score")
	}

	user.Score = score
	return username, score, errors.Wrap(o.db.Update(&user), "could not insert user")
}
