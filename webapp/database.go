// ------------------------------------------------------------------------------------
// This is the 'sqlite' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	"database/sql"
	"errors"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

// ------------------------------------------------------------------------------------
// Our errors related to database operations
// ------------------------------------------------------------------------------------

var UserExistsError = errors.New("named user already exists")

// ------------------------------------------------------------------------------------

type DatabaseStruct struct {
	state                int
	err                  error
	dataSourceName       Path
	Db                   *sql.DB
	theLock              sync.Mutex
	stSelectSpecificBlob *sql.Stmt
	stFindUserIdByName   *sql.Stmt
	blobLock             sync.Mutex
	userLock             sync.Mutex
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

func (db Database) prepareStatements() {
	Todo("have generated functions do this")
	db.stSelectSpecificBlob = db.preparedStatement(`SELECT id FROM ` + tableNameBlob + ` WHERE name = ?`)
	db.stFindUserIdByName = db.preparedStatement(`SELECT id FROM ` + tableNameUser + ` WHERE name = ?`)
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

func (db Database) SetDataSourceName(dataSourceName Path) {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	db.dataSourceName = dataSourceName
	//Alert("<1Setting data source name:", dataSourceName, CurrentDirectory())
}

type ExpObj struct {
	Id     int
	Str    string
	State  UserState
	Amount int
}

func (db Database) Open() error {
	Todo("can probably use generated code for blob table as well, if we support byte arrays")
	Todo("we probably don't need db to cache errors")
	Todo("have generated code accept 'our' db object which wraps the sql database")
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	CheckState(db.dataSourceName.NonEmpty(), "<1No call to SetDataSourceName made")
	// Create the directory containing the database, if it doesn't exist
	dir := db.dataSourceName.Parent().CheckNonEmpty()
	dir.MkDirsM()

	database, err := sql.Open("sqlite3", db.dataSourceName.String())
	db.Db = database
	if db.setError(err) {
		db.state = dbStateFailed
	} else {
		db.state = dbStateOpen

		PrepareDatabase(db.Db)

		// We must create the tables *before* preparing any statements!
		//db.createTables()
		db.prepareStatements()
	}

	if Alert("experiment") {
		u := NewUser().SetEmail("a").SetPassword("pasword").SetState(UserstateActive).SetName("jeff")
		Pr("attempting to read user with id 1")
		uf, errf :=
			ReadUser(db, 1)
		Pr("found user:", uf, "err:", errf)

		Pr("attempting to create user:", INDENT, u)
		u2, err := CreateUser(db, u)
		CheckOk(err)
		Pr("created:", INDENT, u2)
		u3 := u2.ToBuilder().SetName("Frank")
		err2 := UpdateUser(db, u3)
		Pr("updated:", INDENT, u3, "err:", err2)

		Pr("attempting to find user 1 now that it exists")
		uf, errf =
			ReadUser(db, 1)
		Pr("found user:", uf, "err:", errf)

	}

	return db.err
}

func (db Database) Close() error {
	if db.state == dbStateOpen {
		db.Lock()
		defer db.Unlock()
		db.setError(db.Db.Close())
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

const tableNameUser = `user`
const tableNameBlob = `blobtable`

func (db Database) DeleteAllRowsInTable(name string) error {
	db.Lock()
	defer db.Unlock()
	database := db.Db
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

func (db Database) failIfError(err error) {
	if err != nil {
		BadState("<1Serious error has occurred:", err)
	}
}

func (db Database) preparedStatement(sqlStr string) *sql.Stmt {
	st, err := db.Db.Prepare(sqlStr)
	db.failIfError(err)
	return st
}

func (db Database) CreateBlobWithUniqueName(blob []byte) (Blob, error) {

	bb := NewBlob()
	bb.SetData(blob)

	// We use an auxilliary lock to avoid having some other thread call this function
	// and generate the same name (very unlikely)
	db.blobLock.Lock()
	defer db.blobLock.Unlock()

	// Pick a unique blob id (one not already in the blob table)

	pr := PrIf(true)
	pr("choosing unique blob id")
	attempt := 0
	for {
		attempt++
		CheckState(attempt < 50, "failed to choose a unique blob id!")
		bb.SetName(string(GenerateBlobId()))
		pr("blob name:", bb.Name())
		Todo("This code can go away if we get the generated code supporting ReadBlobWith<field>")
		rows := db.stSelectSpecificBlob.QueryRow(bb.Name())
		result := db.scanBlob(rows)
		pr("result:", INDENT, result)
		if result == 0 {
			break
		}
		pr("blob is already in database, attempting again")
	}
	Pr("attempting to insert:", INDENT, bb)
	return CreateBlob(db, bb)
}

func (db Database) scanBlob(rows *sql.Row) int {
	var id int
	err := rows.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		db.setError(err)
	}
	return id
}

// ------------------------------------------------------------------------------------
// User
// ------------------------------------------------------------------------------------

// Create a user with the given (unique) name.

func (db Database) CreateUserByName(user User) (User, error) {

	Todo("Is there a UNIQUENESS constraint that we can take advantage of, to avoid this auxilliary lock?")
	// We use an auxilliary lock to avoid having some other thread call this function
	// and generate the same name (very unlikely)
	db.userLock.Lock()
	defer db.userLock.Unlock()

	var createdUser User

	existingId, _ := ReadUserWithName(db, user.Name())
	Todo("distinguish between a 'no user found' error and some other")
	if existingId != 0 {
		db.setError(UserExistsError)
	} else {
		c, err := CreateUser(db, user)
		createdUser = c
		db.setError(err)
	}

	return createdUser, db.err
}
