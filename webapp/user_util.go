package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	. "github.com/jpsember/golang-base/webserv"
	"strings"
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
	u.SetPassword("password")
	hash, salt := HashPassword(u.Password())
	u.SetPasswordHash(hash)
	u.SetPasswordSalt(salt)

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

func PopulateDatabase(projStruct ProjectStructure) {
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

	// Delete other users
	if !Alert("not deleting other users") {
		delList := []int{}
		iter := UserIterator(0)
		for iter.HasNext() {
			user := iter.Next().(User)
			n := user.Name()
			if strings.HasPrefix(n, "donor") || strings.HasPrefix(n, "admin") || strings.HasPrefix(n, "manager") {
				continue
			}
			Pr("deleting user:", user.Id(), QUO, user.Name())
			delList = append(delList, user.Id())
		}
		for _, id := range delList {
			DeleteUser(id)
		}
	}

	dph := NewDemoPhotos(projStruct.RawPhotosDir(), projStruct.SamplePhotosDir())
	numPhotos := len(dph.ScaledPhotoNames())

	if numPhotos > 0 {
		SamplePhotoBlobIdStart = 2
		for i := 0; i < numPhotos; i++ {
			blobId := i + SamplePhotoBlobIdStart
			bl, _ := ReadBlob(blobId)
			if bl.Id() == 0 {
				// Add a blob from a random photo
				CheckState(blobId > 1, "no placeholder? something strange")
				j := rnd.Intn(numPhotos)
				CreateBlobFromImageFile(dph.ScaledPhotosDir().JoinM(dph.ScaledPhotoNames()[j]))
			}
			SamplePhotoBlobIdCount = blobId + 1 - SamplePhotoBlobIdStart
		}
	}

	for i := 0; i < 100; i++ {
		createAnimalsUpTo(rnd, i+1)
	}
}

var SamplePhotoBlobIdStart int
var SamplePhotoBlobIdCount int

func AttemptSignIn(sess Session, userId int) string {
	pr := PrIf("AttemptSignIn", false)
	var user User
	var prob = ""
	for {
		prob = "No such user, or incorrect password"
		if userId == 0 {
			break
		}

		prob = "User is already logged in"
		if IsUserLoggedIn(userId) {
			break
		}

		prob = "User is unavaliable; sorry"
		user = ReadUserIgnoreError(userId)
		if user.Id() == 0 {
			break
		}

		if AutoActivateUser {
			if user.State() == UserStateWaitingActivation {
				Alert("Activating user automatically (without email verification)")
				user = user.ToBuilder().SetState(UserStateActive).Build()
				UpdateUser(user)
			}
		}

		prob = ""
		switch user.State() {
		case UserStateActive:
			// This is ok.
		case UserStateWaitingActivation:
			prob = "This user has not been activated yet"
		default:
			prob = "This user is in an unsupported state"
		}
		if prob != "" {
			break
		}

		prob = "Unable to log in at this time"
		if !TryLoggingIn(sess, user) {
			break
		}

		prob = ""
		break
	}
	pr("problem is:", prob)
	if prob == "" {
		pr("attempting to select page for user:", INDENT, user)
		switch user.UserClass() {
		case UserClassDonor:
			sess.SwitchToPage(FeedPageTemplate, nil)
			break
		case UserClassManager:
			sess.SwitchToPage(ManagerPageTemplate, nil)
		default:
			NotImplemented("Page for user class:", user.UserClass())
		}
	}
	return prob
}

func DefaultPageForUser(abstractUser AbstractUser) Page {
	if DevGallery {
		return GalleryPageTemplate
	}
	user := abstractUser.(User)
	userId := 0
	if user != nil {
		userId = user.Id()
	}
	var result Page
	if userId == 0 || !IsUserLoggedIn(user.Id()) {
		result = LandingPageTemplate
	} else {
		switch user.UserClass() {
		case UserClassDonor:
			result = FeedPageTemplate
		case UserClassManager:
			result = ManagerPageTemplate
		default:
			NotSupported("page for", user.UserClass())
		}
	}
	return result
}
