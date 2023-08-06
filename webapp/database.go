// ------------------------------------------------------------------------------------
// This is the 'no database' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"reflect"
	"sync"
)

type DatabaseStruct struct {
	state              int
	err                error
	theLock            sync.Mutex
	memTables          map[string]MemTable
	simFilesPath       Path
	changesWrittenTime int64
}

type Database = *DatabaseStruct

const (
	dbStateNew = iota
	dbStateOpen
	dbStateClosed
)

var singletonDatabase Database

func newDatabase() Database {
	t := &DatabaseStruct{}
	t.memTables = make(map[string]MemTable)
	t.changesWrittenTime = CurrentTimeMs()
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

func (db Database) flushChanges() {
	Alert("#50<1Flushing changes")
	for _, mt := range db.memTables {
		if mt.modified {
			p := db.getSimFile(mt)
			p.WriteStringM(mt.table.CompactString())
			mt.modified = false
		}
	}
	db.changesWrittenTime = CurrentTimeMs()
}

// This method does nothing in this version
func (db Database) SetDataSourceName(dataSourceName string) {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
}

func (db Database) Open() {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
	db.state = dbStateOpen
	db.createTables()
}

func (db Database) Close() {
	if db.state == dbStateOpen {
		db.lock()
		defer db.unlock()
		db.flushChanges()
		db.state = dbStateClosed
	}
}

func (db Database) setError(e error) bool {
	if e != nil {
		if db.err != nil {
			db.err = e
			Alert("<1#50Setting database error:", INDENT, e)
		}
	}
	return db.err != nil
}

func (db Database) createTables() {
}

func (db Database) GetAnimal(id int) (Animal, error) {
	db.lock()
	defer db.unlock()
	mp := db.getTable("animal")
	obj := mp.GetData(id, DefaultAnimal)
	return obj.(Animal), db.err
}

func (db Database) AddAnimal(a AnimalBuilder) {
	db.lock()
	defer db.unlock()
	mp := db.getTable("animal")
	id := mp.nextUniqueKey()
	a.SetId(int64(id))
	mp.Put(id, a.Build())
	Todo("write modified table periodically")
	Todo("always writing")
	db.setModified(mp)
}

const SECONDS = 1000
const MINUTES = SECONDS * 60
const HOURS = MINUTES * 60

func (db Database) setModified(mt MemTable) {
	mt.modified = true
	currTime := CurrentTimeMs()
	elapsed := currTime - db.changesWrittenTime
	Pr("setting table modified:", mt.name, "ms since written:", elapsed)
	if elapsed > SECONDS*20 {
		db.flushChanges()
	}
}

func (db Database) flushTable(mt MemTable) {
	p := db.getSimFile(mt)
	p.WriteStringM(mt.table.CompactString())
}

func SQLiteExperiment() {
	Pr("running sim database experiment")

	d := CreateDatabase()
	// We're running from within the webapp subdirectory...
	d.SetDataSourceName("../sqlite/jeff_experiment.db")
	d.Open()

	Pr("opened db")

	for i := 0; i < 100; i++ {
		a := RandomAnimal()
		d.AddAnimal(a)
		Pr("added animal:", INDENT, a)
	}
	d.flushChanges()
}

func (db Database) getTable(name string) MemTable {
	mt := db.memTables[name]
	if mt == nil {
		mt = NewMemTable(name)
		db.memTables[name] = mt
		p := db.getSimFile(mt)
		mt.table = JSMapFromFileIfExistsM(p)
	}
	return mt
}

func (db Database) getSimFile(m MemTable) Path {
	if db.simFilesPath.Empty() {
		db.simFilesPath = NewPathM("simulated_db")
		db.simFilesPath.MkDirsM()
	}
	return db.simFilesPath.JoinM(m.name + ".json")
}

type MemTableStruct struct {
	name     string
	table    JSMap
	modified bool
}

type MemTable = *MemTableStruct

func NewMemTable(name string) MemTable {
	t := &MemTableStruct{
		name:  name,
		table: NewJSMap(),
	}
	return t
}

func (m MemTable) getValue(key string) (JSMap, bool) {
	val, ok := m.table.WrappedMap()[key]
	return val.(JSMap), ok
}

func (m MemTable) nextUniqueKey() int {
	i := 0
	for {
		if !m.table.HasKey(IntToString(i)) {
			Todo("reimplement as binary search for highest key")
			break
		}
		i++
	}
	return i
}

func (m MemTable) GetData(key any, parser DataClass) DataClass {
	strKey := argToMemtableKey(key)
	val := m.table.OptMap(strKey)
	if val == nil {
		return nil
	}
	return parser.Parse(val)
}

func (m MemTable) Put(key any, value any) {
	strKey := argToMemtableKey(key)
	jsmapValue := argToMemtableValue(value)
	m.table.Put(strKey, jsmapValue)
}

func argToMemtableKey(key any) string {
	var strKey string
	switch k := key.(type) {
	case string:
		strKey = k
	case int: // We aren't sure if it's 32 or 64, so choose 64
		strKey = IntToString(k)
	default:
		BadArg("illegal key:", key, "type:", k)
	}
	return strKey
}

func argToMemtableValue(val any) JSMap {
	var strKey JSMap
	switch k := val.(type) {
	case nil:
		break
	case JSMap:
		strKey = k
	default:
		{
			result, ok := val.(DataClass)
			if ok {
				strKey = result.ToJson().AsJSMap()
			}
		}
		break
	}
	if strKey == nil {
		BadArg("illegal value:", val, "type:", reflect.TypeOf(val))
	}
	return strKey
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
