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
	Base         BaseObject
	state        dbState
	err          error
	theLock      sync.Mutex
	memTables    map[string]MemTable
	simFilesPath Path
}

type Database = *DatabaseStruct

type dbState int

const (
	dbStateNew dbState = iota
	dbStateOpen
	dbStateClosed
)

var singletonDatabase Database

func newDatabase() Database {
	t := &DatabaseStruct{}
	t.Base.SetName("Database")
	// t.Base.AlertVerbose()
	t.memTables = make(map[string]MemTable)
	return t
}

func CreateDatabase() Database {
	CheckState(singletonDatabase == nil, "<1Singleton database already exists")
	singletonDatabase = newDatabase()
	b := singletonDatabase.Base
	b.SetName("Database")
	b.AlertVerbose()
	return Db()
}

func Db() Database {
	CheckState(singletonDatabase != nil, "<1No database created yet")
	return singletonDatabase
}

func (db Database) flushChanges() {
	for _, mt := range db.memTables {
		if mt.modified {
			mt.Base.Log("flushing")
			p := db.getSimFile(mt)
			p.WriteStringM(mt.table.CompactString())
			mt.modified = false
		}
	}
}

// This method does nothing in this version
func (db Database) SetDataSourceName(dataSourceName string) {
	CheckState(db.state == dbStateNew, "Illegal state:", db.state)
}

func (db Database) Open() {
	if !db.tryLock(dbStateNew) {
		BadState("Illegal database state")
	}
	defer db.unlock()
	db.state = dbStateOpen
	db.createTables()

	var bgndTask = func() {
		for {
			SleepMs(1000)
			db.Base.Log("flush periodically")
			if !db.tryLock(dbStateOpen) {
				db.Base.Log("...database has closed, exiting")
				return
			}
			db.flushChanges()
			db.theLock.Unlock()
		}
	}
	go bgndTask()
}

func (db Database) Close() {
	if db.tryLock(dbStateOpen) {
		defer db.unlock()
		db.flushChanges()
		db.state = dbStateClosed
	}
}

func (db Database) setError(e error) bool {
	if e != nil {
		if db.err == nil {
			db.err = e
			Alert("<1#50Setting database error:", INDENT, e)
		}
	}
	return db.err != nil
}

const (
	tableNameUser   = "user"
	tableNameAnimal = "animal"
	tableNameBlob   = "blob"
)

func (db Database) createTables() {
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
			break
		}

		mp := db.getTable(tableNameUser)
		key := mp.nextUniqueKey()
		us := user.ToBuilder()
		us.SetId(key)

		createdUser = us.Build()
		mp.Put(us.Id(), createdUser)
		db.setModified(mp)

		CheckState(createdUser.Id() > 0, "unexpected id in new record:", createdUser)
		break
	}

	return createdUser, db.err
}

func (db Database) ReadUser(userId int) (User, error) {
	db.lock()
	defer db.unlock()
	mp := db.getTable(tableNameUser)
	parsedValue := mp.GetData(userId, DefaultUser)
	if parsedValue == nil {
		return nil, nil
	}
	return parsedValue.(User), nil
}

// Write user to database; must already exist.
func (db Database) UpdateUser(user User) error {
	pr := PrIf(false)

	db.lock()
	defer db.unlock()

	for {
		pr("UpdateUser:", INDENT, user)
		mp := db.getTable(tableNameUser)
		if !mp.HasKey(user.Id()) {
			db.setError(ObjectNotFoundError)
			break
		}
		mp.Put(user.Id(), user)
		break
	}
	pr("...returning:", db.err)
	return db.err
}

func (db Database) FindUserWithName(userName string) (int, error) {
	db.lock()
	defer db.unlock()

	foundId := db.auxFindUserWithName(userName)
	if foundId == 0 {
		db.setError(Error("no user with name:", userName))
	}
	return foundId, db.err
}

func (db Database) auxFindUserWithName(userName string) int {
	mp := db.getTable(tableNameUser)
	for id, jsent := range mp.table.WrappedMap() {
		m := jsent.AsJSMap()
		if m.GetString("name") == userName {
			return ParseIntM(id)
		}
	}
	return 0
}

// Write user to database; must already exist.
func (db Database) WriteUser(user User) error {
	db.lock()
	defer db.unlock()
	mp := db.getTable(tableNameUser)
	if !mp.HasKey(user.Id()) {
		return Error("user not found:", user.Id())
	}
	mp.Put(user.Id(), user)
	Todo("have table set itself modified with Put, etc")
	db.setModified(mp)
	return nil
}

// ------------------------------------------------------------------------------------
// Animal
// ------------------------------------------------------------------------------------

func (db Database) AddAnimal(a AnimalBuilder) error {
	db.lock()
	defer db.unlock()
	mp := db.getTable(tableNameAnimal)
	id := mp.nextUniqueKey()
	a.SetId(id)
	mp.Put(id, a.Build())
	db.setModified(mp)
	return db.err
}

func (db Database) ReadAnimal(id int) (Animal, error) {
	db.lock()
	defer db.unlock()
	mp := db.getTable(tableNameAnimal)
	obj := mp.GetData(id, DefaultAnimal)
	return obj.(Animal), db.err
}

const SECONDS = 1000
const MINUTES = SECONDS * 60
const HOURS = MINUTES * 60

func (db Database) setModified(mt MemTable) {
	mt.modified = true
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
	d.Close()
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

func (db Database) getSimDir() Path {
	if db.simFilesPath.Empty() {
		db.simFilesPath = NewPathM("simulated_db")
		db.simFilesPath.MkDirsM()
	}
	return db.simFilesPath
}

func (db Database) getSimFile(m MemTable) Path {
	return db.getSimDir().JoinM(m.name + ".json")
}

type MemTableStruct struct {
	Base     BaseObject
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
	t.Base.SetName("MemTable(" + name + ")")
	t.Base.AlertVerbose()
	return t
}

func (m MemTable) getValue(key string) (JSMap, bool) {
	val, ok := m.table.WrappedMap()[key]
	return val.(JSMap), ok
}

func (m MemTable) nextUniqueKey() int {
	i := 1
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

	defer CatchPanic(func() {
		Pr("Failed to parse:", INDENT, val)
	})

	return parser.Parse(val)
}

func (m MemTable) Put(key any, value any) {
	strKey := argToMemtableKey(key)
	jsmapValue := argToMemtableValue(value)
	m.Base.Log("Writing:", strKey, "=>", INDENT, jsmapValue)
	m.table.Put(strKey, jsmapValue)
}

func (m MemTable) HasKey(key any) bool {
	strKey := argToMemtableKey(key)
	return m.table.HasKey(strKey)
}

func (m MemTable) FindKeyForValue(value any) string {
	mp := m.table.WrappedMap()
	for key, val := range mp {
		if val == value {
			return key
		}
	}
	return ""
}

func argToMemtableKey(key any) string {
	var strKey string
	switch k := key.(type) {
	case string:
		strKey = k
	case int:
		strKey = IntToString(k)
	case int64:
		strKey = IntToString(int(k))
	case int32:
		strKey = IntToString(int(k))
	default:
		BadArg("illegal key:", key, "type:", k, "Info:", Info(key))
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
	if !db.tryLock(dbStateOpen) {
		BadState("<1Illegal state:", db.state)
	}
}

// Attempt to acquire the lock on the database; if state isn't expectedState, releases lock and returns false
func (db Database) tryLock(expectedState dbState) bool {
	db.theLock.Lock()
	if db.state != expectedState {
		db.theLock.Unlock()
		return false
	}
	db.err = nil
	return true
}

func (db Database) unlock() {
	db.theLock.Unlock()
}

func (db Database) failIfError(err error) {
	if err != nil {
		BadState("<1Serious error has occurred:", err)
	}
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
		blobId := GenerateBlobId()
		bb.SetName(blobId.String())
		pr("blob id:", bb.Name())
		p := db.getBlobPath(blobId)
		if !p.Exists() {
			break
		}
		pr("blob is already in database, attempting again")
	}
	Pr("attempting to insert:", INDENT, bb)

	mp := db.getTable(tableNameBlob)
	key := mp.nextUniqueKey()
	bb.SetId(key)
	mp.Put(key, bb.Name())

	db.writeBlob(bb)
	return bb.Build(), nil
}

func (db Database) ReadBlob(blobName BlobId) (Blob, error) {
	//...refer to blobs here by their ids, not their names
	db.lock()
	defer db.unlock()
	pth := db.getBlobPath(blobName)
	if !pth.Exists() {
		return nil, Error("no such blob:", blobName)
	}

	// Find this key in the table
	mp := db.getTable(tableNameBlob)
	blobId := mp.FindKeyForValue(blobName)
	CheckState(blobId != "")

	nameStr := blobName.String()
	bb := NewBlob().SetName(nameStr).SetData(pth.ReadBytesM())

	return bb.Build(), nil
}

func (db Database) getBlobPath(blobId BlobId) Path {
	return db.getSimDir().JoinM(blobId.String() + ".bin")
}

func (db Database) writeBlob(blob Blob) {
	pth := db.getBlobPath(StringToBlobId(blob.Name()))
	pth.WriteBytesM(blob.Data())
}
