// ------------------------------------------------------------------------------------
// This is the 'sqlite' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	"database/sql"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
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
	Todo("Can we use a channel to access the database in a threadsafe manner?")
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

func (db Database) SetError(e error) bool {
	db.err = e
	if db.HasError() {
		Alert("<1#50Setting database error:", INDENT, e)
	}
	return db.HasError()
}

func (db Database) HasError() bool {
	return db.err != nil
}

func (db Database) ClearError() Database {
	db.err = nil
	return db
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
	d.DeleteAllRowsInTable("animal")
	for i := 0; i < 20; i++ {
		a := RandomAnimal()
		d.AddAnimal(a)
		d.AssertOk()
		Pr("added animal:", INDENT, a)
	}
}

func (db Database) CreateTables() {
	database := db.db
	Todo("!Add support for prepared statements")
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
		_, err := database.Exec(create)
		db.SetError(err)
		db.AssertOk()
	}
}

func (db Database) DeleteAllRowsInTable(name string) {
	database := db.db
	Todo("are semicolons needed in sql commands?")
	_, err := database.Exec(`DELETE FROM ` + name)
	db.SetError(err)
	db.AssertOk()
}

func (db Database) AddAnimal(a AnimalBuilder) {
	db.ClearError()
	result, err := db.db.Exec(`INSERT INTO animal (name, summary, details, campaign_target, campaign_balance) VALUES(?,?,?,?,?)`,
		a.Name(), a.Summary(), a.Details(), a.CampaignTarget(), a.CampaignBalance())
	if !db.SetError(err) {
		id, err2 := result.LastInsertId()
		if !db.SetError(err2) {
			a.SetId(id)
		}
	}
}

func (db Database) GetAnimal(id int) Animal {
	Pr("GetAnimal:", id)
	db.ClearError()

	// See https://go.dev/doc/database/prepared-statements

	database := db.db
	const sqlStr string = ` SELECT * FROM animal WHERE uid = ?;`
	stmt, err := database.Prepare(sqlStr)

	db.SetError(err)
	db.AssertOk()

	// Execute the prepared statement, passing in an id value for the
	// parameter whose placeholder is ?

	//var id int
	var name string
	var summary string
	var details string
	var campaignTarget int
	var campaignBalance int

	//name, summary, details, campaign_target, campaign_balance
	rows := stmt.QueryRow(id)
	err = rows.Scan(&id, &name, &summary, &details, &campaignTarget, &campaignBalance)

	if err != nil && err != sql.ErrNoRows {
		db.SetError(err)
	} else if err != sql.ErrNoRows {
		ab := NewAnimal()
		ab.SetId(int64(id))
		ab.SetName(name).SetSummary(summary).SetDetails(details).SetCampaignBalance(int32(campaignBalance)).SetCampaignTarget(int32(campaignTarget))
		Pr("returning:", ab)
		return ab.Build()
	}
	return nil
}
