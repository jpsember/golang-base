package main

import (
	"database/sql"
	. "github.com/jpsember/golang-base/base"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
)

func main() {
	Pr("running db_demo")

	// From https://softchris.github.io/golang-book/05-misc/05-sqlite/

	db := CheckOkWith(sql.Open("sqlite3", "sqlite/jeff_experiment.db"))
	Pr("opened db")

	// Apparently it creates a database if none exists...?

	// Create a table if it doesn't exist
	const create string = `
  CREATE TABLE IF NOT EXISTS user (
  uid INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(64) NOT NULL,
  age INTEGER
  );`

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

	Pr("sleeping...")
	SleepMs(60000)
	Pr("...done sleeping")
}
