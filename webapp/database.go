package webapp

import (
	. "github.com/jpsember/golang-base/base"
	//. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	_ "github.com/mattn/go-sqlite3"
)

// Facade to handle database operations.

type Database interface {
	// Attempt to open the database.  Fails if already open, or previously failed.
	Open()
	CreateTables()
	SetError(error)
}

const (
	DatabaseStateNew = iota
	DatabaseStateOpen
	DatabaseStateClosed
	DatabaseStateFailed
)

var SingletonDatabase Database

func SetSingletonDatabase(db Database) {
	CheckState(SingletonDatabase == nil, "<1Singleton database already exists")
	SingletonDatabase = db
}

func OpenDatabase(db Database) {
	SetSingletonDatabase(db)
	db.Open()
	db.CreateTables()
}
