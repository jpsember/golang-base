// ------------------------------------------------------------------------------------
// This is the 'sqlite' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	"database/sql"
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webapp/gen/webapp_data"
	_ "github.com/mattn/go-sqlite3"
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
	database, err := sql.Open("sqlite3", db.dataSourceName)
	db.db = database
	if db.SetError(err) {
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

func (db Database) AssertOk() Database {
	if db.HasError() {
		BadState("<1DatabaseSqlite has an error:", db.err)
	}
	return db
}

func SQLiteExperiment() {
	Pr("running SQLiteExperiment")

	d := CreateDatabase()
	// We're running from within the webapp subdirectory...
	d.SetDataSourceName("../sqlite/jeff_experiment.db")
	d.Open()
	d.AssertOk()

	Pr("opened db")

	a := RandomAnimal()
	d.AddAnimal(a)
	d.AssertOk()
	Pr("added animal:", INDENT, a)

	//// Apparently it creates a database if none exists...?
	//
	//// Create a table if it doesn't exist
	//const create string = `
	//CREATE TABLE IF NOT EXISTS zebra (
	//uid INTEGER PRIMARY KEY AUTOINCREMENT,
	//name VARCHAR(64) NOT NULL,
	//age INTEGER
	//);`
	//
	//db := d.db
	//
	//CheckOkWith(db.Exec(create))
	//
	//rows := CheckOkWith(db.Query("SELECT * FROM user"))
	//
	//rowTotal := 0
	//for rows.Next() {
	//	rowTotal++
	//	var uid int
	//	var name string
	//	var age int
	//	CheckOk(rows.Scan(&uid, &name, &age))
	//	Pr("uid:", uid, "name:", name, "age:", age)
	//}
	//
	//// I assume this prepares an SQL statement (doing the optimization to determine best way to fulfill the statement)
	//addUserStatement := CheckOkWith(db.Prepare("INSERT INTO user(name, age) values(?,?)"))
	//
	//// If it's empty, create a user
	//if rowTotal == 0 {
	//	res := CheckOkWith(addUserStatement.Exec("Fred", 42))
	//	affected, _ := res.RowsAffected()
	//	Pr("affected rows:", affected)
	//}
	//
	//rnd := rand.New(rand.NewSource(1965))
	//for i := 0; i < 100-rowTotal; i++ {
	//	name := RandomText(rnd, 20, false)
	//	age := rnd.Intn(65) + 8
	//	CheckOkWith(addUserStatement.Exec(name, age))
	//}

}

func (d Database) CreateTables() {
	db := d.db

	if false {
		const drop = `DROP TABLE IF EXISTS animal;`
		_, err := db.Exec(drop)
		d.SetError(err)
		d.AssertOk()
	}

	{
		// Create a table if it doesn't exist
		const create string = `
 CREATE TABLE IF NOT EXISTS animal (
     uid INTEGER PRIMARY KEY AUTOINCREMENT,
     name VARCHAR(64) NOT NULL,
     summary VARCHAR(300),
     details VARCHAR(3000),
     campaign_target INT,
     campaign_balance INT 
     );`
		_, err := db.Exec(create)
		d.SetError(err)
		d.AssertOk()
	}
}

func (d Database) AddAnimal(a webapp_data.AnimalBuilder) {
	d.ClearError()

	result, err := d.db.Exec(`INSERT INTO animal (name, summary, details, campaign_target, campaign_balance) VALUES(?,?,?,?,?)`,
		a.Name(), a.Summary(), a.Details(), a.CampaignTarget(), a.CampaignBalance())
	if !d.SetError(err) {
		id, err2 := result.LastInsertId()
		if !d.SetError(err2) {
			a.SetId(id)
		}
	}

}
