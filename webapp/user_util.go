package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
)

func RandomUser(r JSRand) UserBuilder {
	CheckState(DevDatabase)

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
	CheckState(DevDatabase)

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

func ReadManagers() []User {
	result := []User{}
	iter := UserIterator(0)
	for iter.HasNext() {
		user := iter.Next().(User)
		if user.UserClass() == UserClassManager {
			result = append(result, user)
		}
	}
	return result
}

func createAnimalsUpTo(rnd JSRand, id int) {

	mgrs := ReadManagers()
	CheckState(len(mgrs) != 0)

	rnd = NullToRand(rnd)
	for CheckOkWith(ReadAnimal(id)).Id() == 0 {
		anim := RandomAnimal(rnd, mgrs)
		CreateAnimal(anim)
		Pr("created:", anim.Id(), anim.ManagerId(), anim.Name())
	}
}

func PopulateDatabase() {
	CheckState(DevDatabase)

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

	for i := 0; i < 100; i++ {
		createAnimalsUpTo(rnd, i+1)
	}
}

const (
	UserKeySelectedAnimalId     = "_selected_animal_id"
	UserKeyEditingAnimalImageId = "_editing_animal_image_id"
)
