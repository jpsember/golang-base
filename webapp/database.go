// ------------------------------------------------------------------------------------
// This is the 'no database' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
)

type DatabaseStruct struct {
	state       int
	err         error
	animalTable map[int]webapp_data.Animal
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
	t.animalTable = make(map[int]webapp_data.Animal)
	Todo("read animal table (and others)")
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
func (d Database) SetDataSourceName(dataSourceName string) {
	CheckState(d.state == dbStateNew, "Illegal state:", d.state)
}

func (d Database) Open() {
	CheckState(d.state == dbStateNew, "Illegal state:", d.state)
	d.state = dbStateOpen
	d.CreateTables()
}

func (d Database) Close() {
	d.state = dbStateClosed
}

func (d Database) SetError(e error) bool {
	d.err = e
	if d.HasError() {
		Alert("<1#50Setting database error:", INDENT, e)
	}
	return d.HasError()
}

func (d Database) HasError() bool {
	return d.err != nil
}

func (d Database) ClearError() Database {
	d.err = nil
	return d
}

func (d Database) AssertOk() Database {
	if d.HasError() {
		BadState("<1DatabaseSqlite has an error:", d.err)
	}
	return d
}

func (d Database) CreateTables() {
	Todo("CreateTables")
}

func (d Database) AddAnimal(a webapp_data.AnimalBuilder) {
	mp := d.animalTable
	d.ClearError()
	id := len(mp) + 1
	for HasKey(mp, id) {
		id++
	}

	a.SetId(int64(id))
	mp[id] = a.Build()
	Todo("write modified table periodically")
}
