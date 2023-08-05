// ------------------------------------------------------------------------------------
// This is the 'no database' version of database.go
// ------------------------------------------------------------------------------------

package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"reflect"
)

type DatabaseStruct struct {
	state              int
	err                error
	memTables          map[string]MemTable
	simFilesPath       Path
	changesWrittenTime int64
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

func (d Database) FlushChanges() {
	Alert("#50<1Flushing changes")
	for _, mt := range d.memTables {
		if mt.modified {
			p := d.getSimFile(mt)
			p.WriteStringM(mt.table.CompactString())
			mt.modified = false
		}
	}
	d.changesWrittenTime = CurrentTimeMs()
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
}

func (d Database) GetAnimal(id int) Animal {
	d.ClearError()
	mp := d.getTable("animal")
	obj := mp.GetData(id, DefaultAnimal)
	return obj.(Animal)
}

func (d Database) AddAnimal(a AnimalBuilder) {
	d.ClearError()
	mp := d.getTable("animal")
	id := mp.nextUniqueKey()
	a.SetId(int64(id))
	mp.Put(id, a.Build())
	Todo("write modified table periodically")
	Todo("always writing")
	d.setModified(mp)
}

const SECONDS = 1000
const MINUTES = SECONDS * 60
const HOURS = MINUTES * 60

func (d Database) setModified(mt MemTable) {
	mt.modified = true
	currTime := CurrentTimeMs()
	elapsed := currTime - d.changesWrittenTime
	Pr("setting table modified:", mt.name, "ms since written:", elapsed)
	if elapsed > SECONDS*20 {
		d.FlushChanges()
	}
}

func (d Database) FlushTable(mt MemTable) {
	p := d.getSimFile(mt)
	p.WriteStringM(mt.table.CompactString())
}

func SQLiteExperiment() {
	Pr("running sim database experiment")

	d := CreateDatabase()
	// We're running from within the webapp subdirectory...
	d.SetDataSourceName("../sqlite/jeff_experiment.db")
	d.Open()
	d.AssertOk()

	Pr("opened db")

	for i := 0; i < 100; i++ {
		a := RandomAnimal()
		d.AddAnimal(a)
		d.AssertOk()
		Pr("added animal:", INDENT, a)
	}
	d.FlushChanges()
}

func (d Database) getTable(name string) MemTable {
	mt := d.memTables[name]
	if mt == nil {
		mt = NewMemTable(name)
		d.memTables[name] = mt
		p := d.getSimFile(mt)
		mt.table = JSMapFromFileIfExistsM(p)
	}
	return mt
}

func (d Database) getSimFile(m MemTable) Path {
	if d.simFilesPath.Empty() {
		d.simFilesPath = NewPathM("simulated_db")
		d.simFilesPath.MkDirsM()
	}
	return d.simFilesPath.JoinM(m.name + ".json")
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
