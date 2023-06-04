package main

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/files"
	"os"
)

type ErrLogStruct struct {
	BaseObject
	ownerPath Path
	path      Path
	Errors    int
	Warnings  int
	Clean     bool
}

type ErrLog = *ErrLogStruct

func NewErrLog(ownerPath Path) ErrLog {
	return &ErrLogStruct{
		ownerPath: ownerPath,
	}
}

var Warning = Error("Warning")

func (log ErrLog) Add(err error, messages ...any) error {
	errType := "Error"
	if err == Warning {
		errType = "Warning"
		log.Warnings++
	} else {
		log.Errors++
	}
	errMsg := "*** " + errType + ": " + ToString(messages...)
	Pr(errMsg)

	if log.path.Empty() {
		log.path = NewPathM(log.ownerPath.String() + "_errors.txt")
		if log.Clean && log.path.Exists() {
			log.path.DeleteFileM()
		}
	}

	f, err := os.OpenFile(log.path.String(),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	CheckOk(err, "Failed opening error file:", log.path)
	defer f.Close()
	_, err = f.WriteString(errMsg + "\n")
	CheckOk(err, "Failed appending to error file:", log.path)
	return err
}

func (log ErrLog) PrintSummary() {
	if log.Errors > 0 {
		Pr("*** Total errors:", log.Errors)
	}
	if log.Warnings > 0 {
		Pr("*** Total warnings:", log.Warnings)
	}
	if log.Errors+log.Warnings != 0 {
		Pr("*** See", log.path, "for details.")
	}
}
