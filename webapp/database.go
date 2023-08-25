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
	"strings"
	"sync"
)

// ------------------------------------------------------------------------------------
// Our errors related to database operations
// ------------------------------------------------------------------------------------

var UserExistsError = errors.New("named user already exists")
var UserDoesntExistError = errors.New("user does not exist")
var AnimalDoesntExistError = errors.New("animal does not exist")

// ------------------------------------------------------------------------------------

type DatabaseStruct struct {
	state                  int
	err                    error
	dataSourceName         Path
	db                     *sql.DB
	theLock                sync.Mutex
	stSelectSpecificAnimal *sql.Stmt
	stSelectSpecificBlob   *sql.Stmt
	stSelectSpecificUser   *sql.Stmt
	stFindUserIdByName     *sql.Stmt
	stInsertUser           *sql.Stmt
	stUpdateUser           *sql.Stmt
	stInsertAnimal         *sql.Stmt
	stUpdateAnimal         *sql.Stmt
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
	db.stSelectSpecificAnimal = db.preparedStatement(`SELECT * FROM ` + tableNameAnimal + ` WHERE id = ?`)
	db.stSelectSpecificUser = db.preparedStatement(`SELECT * FROM ` + tableNameUser + ` WHERE id = ?`)
	db.stSelectSpecificBlob = db.preparedStatement(`SELECT * FROM ` + tableNameBlob + ` WHERE id = ?`)
	db.stFindUserIdByName = db.preparedStatement(`SELECT id FROM ` + tableNameUser + ` WHERE name = ?`)
	db.stInsertUser = db.preparedStatement(`INSERT INTO ` + tableNameUser + ` (name, user_state, email, password) VALUES(?,?,?,?)`)
	db.stUpdateUser = db.preparedStatement(`UPDATE ` + tableNameUser + ` SET name = ?, user_state = ?, email = ?, password = ? WHERE id = ?`)
	db.stInsertAnimal = db.preparedStatement(`INSERT INTO ` + tableNameAnimal + ` (name, summary, details, campaign_balance, campaign_target) VALUES(?,?,?,?,?)`)
	db.stUpdateAnimal = db.preparedStatement(`UPDATE ` + tableNameAnimal + ` SET name = ?, summary = ?, details = ?, campaign_balance = ?, campaign_target = ? WHERE id = ?`)
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

func SQLiteExperiment() {
	Pr("running SQLiteExperiment")

	d := CreateDatabase()
	// We're running from within the webapp subdirectory...
	d.SetDataSourceName("../sqlite/animals.db")
	CheckOk(d.Open())

	Pr("opened db")
	CheckOk(d.DeleteAllRowsInTable("animal"))
	for i := 0; i < 20; i++ {
		a := RandomAnimal()
		CheckOkWith(d.CreateAnimal(a))
		Pr("added animal:", INDENT, a)
	}
}

const tableNameAnimal = `animal`
const tableNameUser = `user`
const tableNameBlob = `blobtable`
const tableNameExperiment = `experiment`

func (db Database) createTables() {

	database := db.db

	dbResPath := ProjectDirM().JoinM("webapp/db_res")
	for _, f := range NewDirWalk(dbResPath).IncludeExtensions("txt").Files() {
		if strings.HasSuffix(f.Base(), "_gen") {
			content := f.ReadStringM()
			Todo("do something with content: "+content, f)
			_, err := database.Exec(content)
			db.setError(err)
		}
	}

	{
		var err error

		{
			// Create a table if it doesn't exist
			const create string = `
 CREATE TABLE IF NOT EXISTS ` + tableNameAnimal + ` (
     id INTEGER PRIMARY KEY,
     name VARCHAR(64) NOT NULL,
     summary VARCHAR(300) NOT NULL,
     details VARCHAR(3000) NOT NULL,
     campaign_target INT,
     campaign_balance INT 
     )`
			_, err = database.Exec(create)
			db.setError(err)
		}
		{
			Todo("Use regexp to fill in type stuff, e.g. VARCHAR(20) NOT NULL for password")
			Todo("!Use same limits on user name, email, etc here")
			const create string = `
 CREATE TABLE IF NOT EXISTS ` + tableNameUser + ` (
     id INTEGER PRIMARY KEY,
     name VARCHAR(64) NOT NULL,
     user_state VARCHAR(20) NOT NULL,
	 user_class VARCHAR(20) NOT NULL,
     email VARCHAR(60) NOT NULL,
     password VARCHAR(25) NOT NULL
     )`

			_, err = database.Exec(create)
			db.setError(err)
		}

		{
			_, err = database.Exec(`
CREATE TABLE IF NOT EXISTS ` + tableNameBlob + ` (
    id VARCHAR(36) PRIMARY KEY,
    data BLOB
)`)
			db.setError(err)
		}
	}
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
	db.lock()
	defer db.unlock()

	var createdUser User

	for {
		existingId := db.auxFindUserWithName(user.Name())
		if existingId != 0 {
			db.setError(UserExistsError)
			break
		}

		result, err := db.stInsertUser.Exec(user.Name(), user.State().String(), user.Email(), user.Password())
		if db.setError(err) {
			break
		}

		id, err := result.LastInsertId()
		if db.setError(err) {
			break
		}

		createdUser = user.Build().ToBuilder().SetId(int(id)).Build()
		CheckState(createdUser.Id() > 0, "unexpected id in new record:", createdUser)
		break
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
	pr := PrIf(false)
	pr("ReadUser, id:", userId)

	db.lock()
	defer db.unlock()
	rows := db.stSelectSpecificUser.QueryRow(userId)
	result := db.scanUser(rows)
	pr("result:", INDENT, result, CR, "error:", db.err)
	return result, db.err
}

func (db Database) scanUser(rows *sql.Row) UserBuilder {
	pr := PrIf(false)
	pr("scanUser")

	b := NewUser()

	var id int
	var name string
	var user_state string
	var user_class string
	var email string
	var password string

	errHolder := NewErrorHolder()

	err := rows.Scan(&id, &name, &user_state, &user_class, &email, &password)
	pr("err:", err)
	if err != sql.ErrNoRows {
		errHolder.Add(err)
		b = NewUser()
		b.SetId(id)
		b.SetName(name)
		b.SetState(UserState(UserStateEnumInfo.FromString(user_state, errHolder)))
		b.SetEmail(email)
		b.SetPassword(password)
	}
	db.setError(errHolder.First())
	return b
}

// Write user to database; must already exist.
func (db Database) UpdateUser(user User) error {
	pr := PrIf(false)

	db.lock()
	defer db.unlock()

	for {
		pr("UpdateUser:", INDENT, user)
		result, err := db.stUpdateUser.Exec(user.Name(), user.State().String(), user.Email(), user.Password(), user.Id())
		pr("result:", result, "err:", err)
		if db.setError(err) {
			break
		}

		count, err := result.RowsAffected()
		pr("rows affected:", count, "err:", err)
		if db.setError(err) {
			break
		}

		pr("count:", count)
		if count != 1 {
			db.setError(UserDoesntExistError)
		}
		break
	}
	pr("...returning:", db.err)
	return db.err
}

// ------------------------------------------------------------------------------------
// Animal
// ------------------------------------------------------------------------------------

func (db Database) CreateAnimal(a Animal) (Animal, error) {

	db.lock()
	defer db.unlock()

	var createdAnimal Animal

	for {

		result, err := db.stInsertAnimal.Exec(a.Name(), a.Summary(), a.Details(), a.CampaignBalance(), a.CampaignTarget())
		if db.setError(err) {
			break
		}

		id, err := result.LastInsertId()
		if db.setError(err) {
			break
		}

		createdAnimal = a.Build().ToBuilder().SetId(int(id)).Build()
		CheckState(createdAnimal.Id() > 0, "unexpected id in new record:", createdAnimal)

		break
	}

	return createdAnimal, db.err
}

func (db Database) ReadAnimal(id int) (Animal, error) {
	db.lock()
	defer db.unlock()
	pr := PrIf(false)
	pr("ReadAnimal:", id)
	rows := db.stSelectSpecificAnimal.QueryRow(id)
	result := db.scanAnimal(rows)
	pr("result:", INDENT, result)
	return result, db.err
}

func (db Database) scanAnimal(rows *sql.Row) AnimalBuilder {
	var ab AnimalBuilder
	var id int64
	var name string
	var summary string
	var details string
	var campaignTarget, campaignBalance int32
	err := rows.Scan(&id, &name, &summary, &details, &campaignTarget, &campaignBalance)
	if err != nil && err != sql.ErrNoRows {
		db.setError(err)
	} else if err != sql.ErrNoRows {
		ab = NewAnimal().SetId(int(id)).SetName(name).SetSummary(summary).SetDetails(details).SetCampaignTarget(int(campaignTarget)).SetCampaignBalance(int(campaignBalance))
	}
	return ab
}

// Write animal to database; must already exist.
func (db Database) UpdateAnimal(a Animal) error {
	pr := PrIf(false)

	db.lock()
	defer db.unlock()

	for {
		pr("UpdateAnimal:", INDENT, a)
		Todo("Have convention of listing fields in alpha order here?")
		result, err := db.stUpdateAnimal.Exec(a.Name(), a.Summary(), a.Details(), a.CampaignBalance(), a.CampaignTarget(), a.Id())
		pr("result:", result, "err:", err)
		if db.setError(err) {
			break
		}

		count, err := result.RowsAffected()
		pr("rows affected:", count, "err:", err)
		if db.setError(err) {
			break
		}

		pr("count:", count)
		if count != 1 {
			db.setError(AnimalDoesntExistError)
		}
		break
	}
	pr("...returning:", db.err)
	return db.err
}
