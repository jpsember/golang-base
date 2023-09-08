package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv"
)

func AddDevPageLabel(sess Session, label string) {
	if DevDatabase {
		Alert("?Generating development page labels")
		m := sess.WidgetManager()

		user := OptSessionUser(sess)
		if user.Id() != 0 {
			label = label + ", user:" + user.Name()
		}
		m.Size(SizeMicro).Align(AlignRight).Label(label).AddHeading()
	}
}
