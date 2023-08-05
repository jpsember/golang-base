// ------------------------------------------------------------------------------------
// This is the 'sqlite' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	"database/sql"
	. "github.com/jpsember/golang-base/base"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
)

type DatabaseStruct struct {
	state          int
	err            error
	dataSourceName string
	db             *sql.DB
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

func (db Database) SetDataSourceName(dataSourceName string) {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	db.dataSourceName = dataSourceName
	Alert("<1Setting data source name:", dataSourceName, CurrentDirectory())
}

func (db Database) Open() {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	CheckState(db.dataSourceName != "", "<1No call to SetDataSourceName made")
	db.db, db.err = sql.Open("sqlite3", db.dataSourceName)
	if db.ErrorOccurred() {
		db.state = dbStateFailed
		return
	}
	db.state = dbStateOpen
	db.CreateTables()
}

func (db Database) Close() {
	if db.state == dbStateOpen {
		db.err = db.db.Close()
		db.state = dbStateClosed
	}
}

func (d Database) SetError(e error) {
	d.err = e
	if e != nil {
		Alert("<1#50Setting database error:", INDENT, e)
	}
}

func (db Database) AssertOk() Database {
	if db.err != nil {
		BadState("<1DatabaseSqlite has an error:", db.err)
	}
	return db
}

func (db Database) ErrorOccurred() bool {
	if db.err != nil {
		Pr("*** Database error occurred:", INDENT, db.err)
		return true
	}
	return false
}

func SQLiteExperiment() {
	Pr("running SQLiteExperiment")

	d := CreateDatabase()
	// We're running from within the webapp subdirectory...
	d.SetDataSourceName("../sqlite/jeff_experiment.db")
	d.Open()
	d.AssertOk()

	Pr("opened db")

	// Apparently it creates a database if none exists...?

	// Create a table if it doesn't exist
	const create string = `
  CREATE TABLE IF NOT EXISTS zebra (
  uid INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(64) NOT NULL,
  age INTEGER
  );`

	db := d.db

	CheckOkWith(db.Exec(create))

	rows := CheckOkWith(db.Query("SELECT * FROM user"))

	rowTotal := 0
	for rows.Next() {
		rowTotal++
		var uid int
		var name string
		var age int
		CheckOk(rows.Scan(&uid, &name, &age))
		Pr("uid:", uid, "name:", name, "age:", age)
	}

	// I assume this prepares an SQL statement (doing the optimization to determine best way to fulfill the statement)
	addUserStatement := CheckOkWith(db.Prepare("INSERT INTO user(name, age) values(?,?)"))

	// If it's empty, create a user
	if rowTotal == 0 {
		res := CheckOkWith(addUserStatement.Exec("Fred", 42))
		affected, _ := res.RowsAffected()
		Pr("affected rows:", affected)
	}

	rnd := rand.New(rand.NewSource(1965))
	for i := 0; i < 100-rowTotal; i++ {
		name := RandomText(rnd, 20, false)
		age := rnd.Intn(65) + 8
		CheckOkWith(addUserStatement.Exec(name, age))
	}

}

func (d Database) CreateTables() {
	db := d.db
	// Create a table if it doesn't exist
	const create string = `
 CREATE TABLE IF NOT EXISTS animal (
     uid INTEGER PRIMARY KEY AUTOINCREMENT,
     name VARCHAR(64) NOT NULL,
     summary VARCHAR(300),
     details VARCHAR(3000),
     campaign_target INT,
     campain_balance INT 
     );`
	_, err := db.Exec(create)
	d.SetError(err)
	d.AssertOk()
}
