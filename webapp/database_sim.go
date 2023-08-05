package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

type DatabaseSimStruct struct {
	state int
	err   error
}

type DatabaseSim = *DatabaseSimStruct

func NewDatabaseSim() DatabaseSim {
	t := &DatabaseSimStruct{}
	return t
}

// Verify that DatabaseSim implements the Database interface
var _ Database = (DatabaseSim)(nil)

func (db DatabaseSim) Open() {
	CheckState(db.state == DatabaseStateNew, "Illegal state:", db.state)
	db.state = DatabaseStateOpen
}

func (db DatabaseSim) Close() {
	db.state = DatabaseStateClosed
}

func (d DatabaseSim) SetError(e error) {
	d.err = e
	if e != nil {
		Pr("*** Setting database error:", INDENT, e)
	}
}

func (d DatabaseSim) CreateTables() {
	Todo("CreateTables")
}
