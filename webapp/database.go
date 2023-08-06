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
	state                    int
	err                      error
	dataSourceName           string
	db                       *sql.DB
	theLock                  sync.Mutex
	stmtSelectSpecificAnimal *sql.Stmt
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
	db.stmtSelectSpecificAnimal = db.preparedStatement(`SELECT * FROM animal WHERE uid = ?`)
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

func (db Database) SetDataSourceName(dataSourceName string) {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	db.dataSourceName = dataSourceName
	Alert("<1Setting data source name:", dataSourceName, CurrentDirectory())
}

func (db Database) Open() error {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	CheckState(db.dataSourceName != "", "<1No call to SetDataSourceName made")
	database, err := sql.Open("sqlite3", db.dataSourceName)
	db.db = database
	if db.setError(err) {
		db.state = dbStateFailed
	} else {
		db.state = dbStateOpen
		db.prepareStatements()
		db.createTables()
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
		if db.err != nil {
			db.err = err
			Alert("<1#50Setting database error:", INDENT, err)
		}
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

	Pr("opened db")
	d.DeleteAllRowsInTable("animal")
	for i := 0; i < 20; i++ {
		a := RandomAnimal()
		d.AddAnimal(a)
		Pr("added animal:", INDENT, a)
	}
}

func (db Database) createTables() {
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
		db.setError(err)
	}
}

func (db Database) DeleteAllRowsInTable(name string) error {
	db.lock()
	defer db.unlock()
	database := db.db
	Todo("are semicolons needed in sql commands?")
	_, err := database.Exec(`DELETE FROM ` + name)
	db.setError(err)
	return db.err
}

func (db Database) AddAnimal(a AnimalBuilder) error {
	db.lock()
	defer db.unlock()
	result, err := db.db.Exec(`INSERT INTO animal (name, summary, details, campaign_target, campaign_balance) VALUES(?,?,?,?,?)`,
		a.Name(), a.Summary(), a.Details(), a.CampaignTarget(), a.CampaignBalance())
	if !db.setError(err) {
		id, err2 := result.LastInsertId()
		if !db.setError(err2) {
			a.SetId(id)
		}
	}
	return db.err
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
		ab = NewAnimal().SetId(id).SetName(name).SetSummary(summary).SetDetails(details).SetCampaignTarget(campaignTarget).SetCampaignBalance(campaignBalance)
	}
	return ab
}

func (db Database) GetAnimal(id int) (Animal, error) {
	db.lock()
	defer db.unlock()
	pr := PrIf(false)
	pr("GetAnimal:", id)
	rows := db.stmtSelectSpecificAnimal.QueryRow(id)
	result := db.scanAnimal(rows)
	pr("result:", INDENT, result)
	return result, db.err
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
	stmt, err := db.db.Prepare(sqlStr)
	db.failIfError(err)
	return stmt
}
