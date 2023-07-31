package webserv

import (
	"database/sql"
	. "github.com/jpsember/golang-base/base"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

type DbSessionManager struct {
	BaseObject
	db                   *sql.DB
	lock                 sync.RWMutex
	persistPath          Path
	addSessionStatement  *sql.Stmt
	findSessionStatement *sql.Stmt
}

const TableNameSession = "session"

const sessionIdLength = 16 // 16 base64 characters, so (2^(16*8)) possible sessions, a huge number; but is hijacking still a possibility?

func BuildDbSessionManager(db *sql.DB) *DbSessionManager {
	sm := new(DbSessionManager)
	sm.SetName("DbSessionManager")
	sm.db = db

	// Create a table if it doesn't exist
	create := `
  CREATE TABLE IF NOT EXISTS session (
  session_id CHAR(` + IntToString(sessionIdLength) + `) PRIMARY KEY,
  user_id INTEGER,
  creation_time INTEGER
  ) STRICT;`

	CheckOkWith(db.Exec(create))

	// Prepare some statements
	sm.addSessionStatement = CheckOkWith(db.Prepare("INSERT INTO session(session_id, creation_time) values(?,?)"))
	sm.findSessionStatement = CheckOkWith(db.Prepare("SELECT * FROM session WHERE session_id = ?"))
	return sm
}

func (s *DbSessionManager) SetModified(session Session) {
}

func (s *DbSessionManager) FindSession(id string) Session {
	s.Log("FindSession, id:", id)

	Todo("ensure id has expected length before calling db?")
	res := CheckOkWith(s.findSessionStatement.Query(id))
	if res.Next() {
		var found_id string
		var user_id int64
		var creation_time int64
		CheckOk(res.Scan(&found_id, &user_id, &creation_time))
		Pr("found id:", found_id, "user_id", user_id, "creation_time", creation_time)
		CheckState(found_id == id, "found_id", found_id, "not equal to", id)

	}
	return nil
}
