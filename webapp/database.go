package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

// Facade to handle database operations.

type Database interface {
	// Attempt to open the database.  Fails if already open, or previously failed.
	Open()
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

func OpenDatabase() {
	db := NewDatabaseSim()
	SetSingletonDatabase(db)
}
