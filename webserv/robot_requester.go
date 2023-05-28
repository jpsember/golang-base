package webserv

import (
	. "github.com/jpsember/golang-base/base"
	"io"
	"net/http"
	"time"
)

type RobotRequester struct {
	BaseObject
	IntervalMS int
	url        string
	ticker     *time.Ticker
}

func NewRobotRequester(url string) *RobotRequester {
	r := &RobotRequester{
		url:        url,
		IntervalMS: 5000,
	}
	r.SetName("RobotRequester")
	return r
}

func (r *RobotRequester) Start() {
	r.ticker = time.NewTicker(time.Duration(r.IntervalMS) * time.Millisecond)
	Todo("How do we stop this thing?")
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.makeRequest()
			case <-quit:
				r.ticker.Stop()
				r.ticker = nil
				return
			}
		}
	}()
}

func (r *RobotRequester) makeRequest() {
	var err error
	for {
		resp, err2 := http.Get(r.url)
		err = err2
		if err != nil {
			break
		}

		resBody, err := io.ReadAll(resp.Body)
		err = err2
		if err != nil {
			break
		}
		r.Log("received:", INDENT, string(resBody))
		break
	}
	if err != nil {
		r.Log("Problem with request:", err)
	}
}
