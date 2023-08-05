// ------------------------------------------------------------------------------------
// This is the 'no database' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
)

type DatabaseStruct struct {
	state int
	err   error
}

type Database = *DatabaseStruct

const (
	dbStateNew = iota
	dbStateOpen
	dbStateClosed
	dbStateFailed
)

var singletonDatabase Database

func newDatabase() Database {
	t := &DatabaseStruct{}
	return t
}

func CreateDatabase() Database {
	CheckState(singletonDatabase == nil, "<1Singleton database already exists")
	singletonDatabase = newDatabase()
	return Db()
}

func Db() Database {
	CheckState(singletonDatabase != nil, "<1No database created yet")
	return singletonDatabase
}

// This method does nothing in this version
func (db Database) SetDataSourceName(dataSourceName string) {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
}

func (db Database) Open() {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	db.state = dbStateOpen
	db.CreateTables()
}

func (db Database) Close() {
	db.state = dbStateClosed
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

func (d Database) AddAnimal(a webapp_data.AnimalBuilder) {
	//d.ClearError()
	//
	//result, err := d.db.Exec(`INSERT INTO animal (name, summary, details, campaign_target, campaign_balance) VALUES(?,?,?,?,?)`,
	//	a.Name(), a.Summary(), a.Details(), a.CampaignTarget(), a.CampaignBalance())
	//if !d.SetError(err) {
	//	id, err2 := result.LastInsertId()
	//	if !d.SetError(err2) {
	//		a.SetId(id)
	//	}
	//}
	NotImplemented()
}
