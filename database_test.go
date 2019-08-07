package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestUserInsert(t *testing.T) {
	db, cleanup := setup()
	defer cleanup()

	data := []struct {
		name  string
		score int
		err   error
	}{
		{
			name:  "george.abitbol",
			score: 42,
			err:   nil,
		},
		{
			name:  "george.abitbol",
			score: 42,
			err:   storm.ErrAlreadyExists,
		},
		{
			name:  "george.abitbol.2",
			score: 42,
			err:   nil,
		},
	}

	for _, d := range data {
		name, score, err := db.Insert(d.name, d.score)

		assert.Equal(t, d.err, errors.Cause(err))
		assert.Equal(t, d.name, name)
		assert.Equal(t, d.score, score)
	}
}

func TestUserUpdate(t *testing.T) {
	db, cleanup := setup()
	defer cleanup()

	name, score, err := db.Update("george.abitbol", 42)
	assert.Equal(t, storm.ErrNotFound, errors.Cause(err))
	assert.Equal(t, "george.abitbol", name)
	assert.Equal(t, 42, score)

	//

	_, _, err = db.Insert("george.abitbol", 42)
	assert.NoError(t, err)

	//

	name, score, err = db.Update("george.abitbol", 4242)
	assert.NoError(t, err)
	assert.Equal(t, "george.abitbol", name)
	assert.Equal(t, 4242, score)
}

func TestUserExists(t *testing.T) {
	db, cleanup := setup()
	defer cleanup()

	ok, err := db.Exists("george.abitbol")
	assert.NoError(t, err)
	assert.False(t, ok)

	//

	_, _, err = db.Insert("george.abitbol", 42)
	assert.NoError(t, err)

	//

	ok, err = db.Exists("george.abitbol")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func setup() (db Database, cleanup func()) {
	tmpfile, err := ioutil.TempFile("", "rpi.*.db")
	if err != nil {
		panic(err)
	}
	filename := tmpfile.Name()
	tmpfile.Close()

	db, err = Open(filename)
	if err != nil {
		panic(err)
	}

	return db, func() {
		db.Close()
		os.RemoveAll(filename)
	}
}
