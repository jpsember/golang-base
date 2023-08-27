// ------------------------------------------------------------------------------------
// This is the 'sqlite' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	"database/sql"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

type DatabaseStruct struct {
	state          int
	err            error
	dataSourceName Path
	sqlDatabase    *sql.DB
	theLock        sync.Mutex
	blobLock       sync.Mutex
	userLock       sync.Mutex
}

type Database = *DatabaseStruct

const (
	dbStateNew = iota
	dbStateOpen
	dbStateClosed
	dbStateFailed
)

func newDatabase() Database {
	t := &DatabaseStruct{}
	return t
}

var singletonDatabase Database

func CreateDatabase() Database {
	CheckState(singletonDatabase == nil, "<1Singleton database already exists")
	singletonDatabase = newDatabase()
	return Db()
}

func Db() Database {
	CheckState(singletonDatabase != nil, "<1No database created yet")
	return singletonDatabase
}

func (db Database) SetDataSourceName(dataSourceName Path) {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	db.dataSourceName = dataSourceName
	//Alert("<1Setting data source name:", dataSourceName, CurrentDirectory())
}

func (db Database) Open() error {
	Todo("we probably don't need db to cache errors")
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	CheckState(db.dataSourceName.NonEmpty(), "<1No call to SetDataSourceName made")
	// Create the directory containing the database, if it doesn't exist
	dir := db.dataSourceName.Parent().CheckNonEmpty()
	dir.MkDirsM()

	database, err := sql.Open("sqlite3", db.dataSourceName.String())
	db.sqlDatabase = database
	if db.setError(err) {
		db.state = dbStateFailed
	} else {
		db.state = dbStateOpen
		PrepareDatabase(db.sqlDatabase, &db.theLock)
	}

	if Alert("experiment with blob") {
		var dat = []byte{1, 2, 3, 4}

		bb := NewBlob()
		bb.SetData(dat)

		key1 := string(GenerateBlobId())

		bb.SetName(key1)
		result, err := CreateBlobWithName(bb)
		CheckOk(err)

		bb = result.ToBuilder()
		result2, err2 := CreateBlobWithName(bb)
		Pr("attempt to create duplicate blob:", result2, "err:", err2)
	}
	if !Alert("experiment") {
		u := NewUser().SetEmail("a").SetPassword("pasword").SetState(UserstateActive).SetName("jeff")
		Pr("attempting to read user with id 1")
		uf, errf :=
			ReadUser(1)
		Pr("found user:", uf, "err:", errf)

		Pr("attempting to create user:", INDENT, u)
		u2, err := CreateUser(u)
		CheckOk(err)
		Pr("created:", INDENT, u2)
		u3 := u2.ToBuilder().SetName("Frank")
		err2 := UpdateUser(u3)
		Pr("updated:", INDENT, u3, "err:", err2)

		Pr("attempting to find user 1 now that it exists")
		uf, errf =
			ReadUser(1)
		Pr("found user:", uf, "err:", errf)
	}

	return db.err
}

func (db Database) Close() error {
	if db.state == dbStateOpen {
		db.Lock()
		defer db.Unlock()
		db.setError(db.sqlDatabase.Close())
		db.state = dbStateClosed
	}
	return db.err
}

// If no registered error exists, set it.  Return true if registered error exists afterwards.
func (db Database) setError(err error) bool {
	if err != nil {
		if db.err == nil {
			db.err = err
			Alert("<1#50Setting database error:", INDENT, err)
		}
		return true
	}
	return false
}

func (db Database) ok() bool {
	return db.err == nil
}

func (db Database) DeleteAllRowsInTable(name string) error {
	db.Lock()
	defer db.Unlock()
	database := db.sqlDatabase
	_, err := database.Exec(`DELETE FROM ` + name)
	db.setError(err)
	return db.err
}

// Acquire the lock on the database, and clear the error register.
func (db Database) Lock() {
	if db.state != dbStateOpen {
		BadState("<1Illegal state:", db.state)
	}
	db.theLock.Lock()
	db.err = nil
}

func (db Database) Unlock() {
	db.theLock.Unlock()
}
