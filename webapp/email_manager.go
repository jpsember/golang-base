package webapp

import (
	. "github.com/jpsember/golang-base/base"
	"github.com/jpsember/golang-base/webserv"
	. "github.com/jpsember/golang-base/webserv/gen/webserv_data"
	"sync"
)

type EmailManagerStruct struct {
	config           ZohoConfig
	lock             sync.Mutex
	pendingEmails    []Email
	emailQueue       []Email
	actualEmailsSent int
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

	// Try sending queued emails

	var newQueue []Email
	for _, email := range m.emailQueue {
		if Alert("limiting actual emails sent") && m.actualEmailsSent >= 1 {
			continue
		}
		err := webserv.SharedZoho().SendEmail(email)
		if err != nil {
			ReportIfError(err, "failed to send email:", INDENT, email)
			newQueue = append(newQueue, email)
			continue
		}
		m.actualEmailsSent++
	}
	m.emailQueue = newQueue

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
