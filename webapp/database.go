package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

// ------------------------------------------------------------------------------------
// What follows is the 'no database' version of the code
// ------------------------------------------------------------------------------------

type DatabaseStruct struct {
	state int
	err   error
}

type Database = *DatabaseStruct

const (
	DatabaseStateNew = iota
	DatabaseStateOpen
	DatabaseStateClosed
	DatabaseStateFailed
)

var SingletonDatabase Database

func CreateDatabase() Database {
	CheckState(SingletonDatabase == nil, "<1Singleton database already exists")
	SingletonDatabase = newDatabase()
	return Db()
}

func Db() Database {
	CheckState(SingletonDatabase != nil, "<1No database created yet")
	return SingletonDatabase
}

func newDatabase() Database {
	t := &DatabaseStruct{}
	return t
}

func (db Database) Open() {
	CheckState(db.state == DatabaseStateNew, "Illegal state:", db.state)
	db.state = DatabaseStateOpen
	db.CreateTables()
}

func (db Database) Close() {
	db.state = DatabaseStateClosed
}

func (d Database) SetError(e error) {
	d.err = e
	if e != nil {
		Pr("*** Setting database error:", INDENT, e)
	}
}

func (d Database) CreateTables() {
	Todo("CreateTables")
}
