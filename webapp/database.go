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
var UserDoesntExistError = errors.New("user does not exist")

// ------------------------------------------------------------------------------------

type DatabaseStruct struct {
	state                int
	err                  error
	dataSourceName       Path
	db                   *sql.DB
	theLock              sync.Mutex
	stSelectSpecificBlob *sql.Stmt
	stFindUserIdByName   *sql.Stmt
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
	db.stSelectSpecificBlob = db.preparedStatement(`SELECT * FROM ` + tableNameBlob + ` WHERE id = ?`)
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
	db.db = database
	if db.setError(err) {
		db.state = dbStateFailed
	} else {
		db.state = dbStateOpen
		// We must create the tables *before* preparing any statements!
		db.createTables()
		db.prepareStatements()
	}

	if Alert("experiment") {
		u := NewUser().SetEmail("a").SetPassword("pasword").SetState(UserstateActive).SetName("jeff")
		Pr("attempting to read user with id 1")
		uf, errf :=
			ReadUser(db.db, 1)
		Pr("found user:", uf, "err:", errf)

		Pr("attempting to create user:", INDENT, u)
		u2, err := CreateUser(db.db, u)
		CheckOk(err)
		Pr("created:", INDENT, u2)
		u3 := u2.ToBuilder().SetName("Frank")
		err2 := UpdateUser(db.db, u3)
		Pr("updated:", INDENT, u3, "err:", err2)

		Pr("attempting to find user 1 now that it exists")
		uf, errf =
			ReadUser(db.db, 1)
		Pr("found user:", uf, "err:", errf)

	}

	return db.err
}

func (db Database) Close() error {
	if db.state == dbStateOpen {
		db.lock()
		defer db.unlock()
		db.setError(db.db.Close())
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

const tableNameAnimal = `animal`
const tableNameUser = `user`
const tableNameBlob = `blobtable`

func (db Database) createTables() {

	database := db.db

	CreateTableUser(database)
	CreateTableAnimal(database)

	_, err := database.Exec(`
CREATE TABLE IF NOT EXISTS ` + tableNameBlob + ` (
    id VARCHAR(36) PRIMARY KEY,
    data BLOB
)`)
	db.setError(err)

}

func (db Database) DeleteAllRowsInTable(name string) error {
	db.lock()
	defer db.unlock()
	database := db.db
	_, err := database.Exec(`DELETE FROM ` + name)
	db.setError(err)
	return db.err
}

// Acquire the lock on the database, and clear the error register.
func (db Database) lock() {
	if db.state != dbStateOpen {
		BadState("<1Illegal state:", db.state)
	}
	db.theLock.Lock()
	db.err = nil
}

func (db Database) unlock() {
	db.theLock.Unlock()
}

func (db Database) failIfError(err error) {
	if err != nil {
		BadState("<1Serious error has occurred:", err)
	}
}

func (db Database) preparedStatement(sqlStr string) *sql.Stmt {
	st, err := db.db.Prepare(sqlStr)
	db.failIfError(err)
	return st
}

func (db Database) InsertBlob(blob []byte) (Blob, error) {
	db.lock()
	defer db.unlock()

	bb := NewBlob()
	bb.SetData(blob)

	// Pick a unique blob id (one not already in the blob table)

	pr := PrIf(true)
	pr("choosing unique blob id")
	attempt := 0
	for {
		attempt++
		CheckState(attempt < 50, "failed to choose a unique blob id!")
		bb.SetId(string(GenerateBlobId()))
		pr("blob id:", bb.Id())
		rows := db.stSelectSpecificBlob.QueryRow(bb.Id())
		result := db.scanBlob(rows)
		pr("result:", INDENT, result)
		if result == nil {
			break
		}
		pr("blob is already in database, attempting again")
	}
	Pr("attempting to insert:", INDENT, bb)

	_, err := db.db.Exec(`INSERT INTO `+tableNameBlob+` (id, data) VALUES(?,?)`, bb.Id(), bb.Data())
	return bb.Build(), err
}

func (db Database) ReadBlob(blobId BlobId) (Blob, error) {
	db.lock()
	defer db.unlock()

	idStr := blobId
	rows := db.stSelectSpecificBlob.QueryRow(idStr)
	bb := db.scanBlob(rows)
	var b Blob
	if db.ok() {
		b = bb.Build()
	}
	return b, db.err
}

func (db Database) scanBlob(rows *sql.Row) BlobBuilder {
	var ab BlobBuilder
	var id string
	var data []byte
	err := rows.Scan(&id, &data)
	if err != nil && err != sql.ErrNoRows {
		db.setError(err)
	} else if err != sql.ErrNoRows {
		ab = NewBlob().SetId(id).SetData(data)
	}
	return ab
}

// ------------------------------------------------------------------------------------
// User
// ------------------------------------------------------------------------------------

// Create a user with the given (unique) name.
func (db Database) CreateUser(user User) (User, error) {
	Todo("maybe put all this boilerplat (lock/unlock) within generated code")
	db.lock()
	defer db.unlock()

	var createdUser User

	existingId := db.auxFindUserWithName(user.Name())
	if existingId != 0 {
		db.setError(UserExistsError)

	} else {
		c, err := CreateUser(db.db, user)
		createdUser = c
		db.setError(err)
	}

	return createdUser, db.err
}

func (db Database) FindUserWithName(userName string) (int, error) {
	pr := PrIf(false)
	pr("FindUserWithName:", userName)

	db.lock()
	defer db.unlock()

	foundId := db.auxFindUserWithName(userName)
	if foundId == 0 {
		db.setError(UserDoesntExistError)
	}
	pr("returning foundId", foundId, "error", db.err)

	return foundId, db.err
}

func (db Database) auxFindUserWithName(userName string) int {
	rows := db.stFindUserIdByName.QueryRow(userName)
	var id int
	err := rows.Scan(&id)
	if err != sql.ErrNoRows {
		db.setError(err)
	}
	return id
}

func (db Database) ReadUser(userId int) (User, error) {
	db.lock()
	defer db.unlock()
	return ReadUser(db.db, userId)
}

// Write user to database; must already exist.
func (db Database) UpdateUser(user User) error {
	pr := PrIf(false)

	db.lock()
	defer db.unlock()

	db.setError(UpdateUser(db.db, user))

	pr("...returning:", db.err)
	return db.err
}

// ------------------------------------------------------------------------------------
// Animal
// ------------------------------------------------------------------------------------

func (db Database) CreateAnimal(a Animal) (Animal, error) {

	db.lock()
	defer db.unlock()

	createdAnimal, err := CreateAnimal(db.db, a)
	db.setError(err)
	return createdAnimal, db.err
}

func (db Database) ReadAnimal(id int) (Animal, error) {
	db.lock()
	defer db.unlock()
	return ReadAnimal(db.db, id)
}

// Write animal to database; must already exist.
func (db Database) UpdateAnimal(a Animal) error {
	db.lock()
	defer db.unlock()

	db.setError(UpdateAnimal(db.db, a))
	return db.err
}
