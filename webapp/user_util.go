package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
)

func RandomUser(r JSRand) UserBuilder {
	r = NullToRand(r)
	a := NewUser()
	a.SetEmail(RandomEmailAddress(r))
	a.SetName("Donor" + RandomWord(r))
	a.SetUserClass(UserClassDonor)
	a.SetPassword("password")
	a.SetState(UserStateActive)
	return a
}

func HasUsers() bool {
	return CheckOkWith(ReadUser(1)) != DefaultUser
}

func GenerateRandomUsers() {
	Alert("<1Generating some random users")
	rnd := NewJSRand()

	for i := 0; i < 8; i++ {
		user := RandomUser(rnd)
		CreateUser(user)
		Pr("added user:", INDENT, user)
	}
}

func createUserIfMissing(name string, class UserClass) {
	if CheckOkWith(ReadUserWithName(name)).Id() != 0 {
		return
	}
	u := NewUser().SetName(name).SetUserClass(class)
	CheckOkWith(CreateUser(u))
}

func createAnimalsUpTo(rnd JSRand, id int) {
	rnd = NullToRand(rnd)
	for CheckOkWith(ReadAnimal(id)).Id() == 0 {
		anim := RandomAnimal(rnd)
		CreateAnimal(anim)
	}
}

func PopulateDatabase() {

	Alert("!<1Populating database if empty")
	rnd := NewJSRand().SetSeed(1965)

	for i := 0; i < 8; i++ {
		createUserIfMissing("donor"+IntToString(i+1), UserClassDonor)
	}
	for i := 0; i < 2; i++ {
		createUserIfMissing("admin"+IntToString(i+1), UserClassAdmin)
	}
	for i := 0; i < 5; i++ {
		createUserIfMissing("manager"+IntToString(i+1), UserClassManager)
	}

	for i := 0; i < 30; i++ {
		createAnimalsUpTo(rnd, i+1)
	}
}
