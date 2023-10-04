package webapp

import (
	. "github.com/jpsember/golang-base/base"
)

type EmailManagerStruct struct {
}

type EmailManager = *EmailManagerStruct

var sharedEmailManager EmailManager

func SharedEmailManager() EmailManager {
	if sharedEmailManager == nil {
		t := &EmailManagerStruct{}
		sharedEmailManager = t
	}
	return sharedEmailManager
}

func (m EmailManager) Start() {
	go m.backgroundTask()
}

func (m EmailManager) backgroundTask() {
	pr := PrIf("EmailManager.backgroundTask", true)
	pr("starting")
	for {
		SleepMs(5000)
		pr("tick")
	}
}
