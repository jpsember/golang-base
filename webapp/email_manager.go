package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webserv"
	. "github.com/jpsember/golang-base/webserv/gen/webserv_data"
	"sync"
)

type EmailManagerStruct struct {
	config        ZohoConfig
	lock          sync.Mutex
	pendingEmails []Email
	emailQueue    []Email
}

type EmailManager = *EmailManagerStruct

var sharedEmailManager EmailManager

func SharedEmailManager() EmailManager {
	if sharedEmailManager == nil {
		t := &EmailManagerStruct{}
		sharedEmailManager = t
		t.config = webserv.SharedZoho().Config()
	}
	return sharedEmailManager
}

func (m EmailManager) Start() {
	go m.backgroundTask()
}

func (m EmailManager) backgroundTask() {
	pr := PrIf("EmailManager.backgroundTask", false)
	pr("starting")
	for {
		SleepMs(5000)
		pr("tick")
		m.backgroundIter()
	}
}

func (m EmailManager) backgroundIter() {
	pr := PrIf("EmailManager.backgroundIter", false)
	// Move any accumulated emails from the public queue to our internal one
	{
		m.lock.Lock()
		m.emailQueue = append(m.emailQueue, m.pendingEmails...)
		m.pendingEmails = nil
		m.lock.Unlock()
	}
	pr("pending emails count:", len(m.pendingEmails))
}

func (m EmailManager) Send(email Email) {
	pr := PrIf("EmailManager.Send", true)
	m.lock.Lock()
	defer m.lock.Unlock()
	m.pendingEmails = append(m.pendingEmails, email)
	pr("sending:", INDENT, email)
	pr("pending emails now:", INDENT, m.pendingEmails)

}
