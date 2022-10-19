package qblog

import "log"

var (
	Log = new(QBLog)
)

type QBLog struct {
	qbDebug bool
}

func (slf *QBLog) SetDebug(b bool) {
	slf.qbDebug = b
}

func (slf *QBLog) Debug(a ...interface{}) {
	if slf.qbDebug {
		log.Println("DEBUG:", a)
	}
}

func (slf *QBLog) Info(a ...interface{}) {
	log.Println("INFO:", a)
}

func (slf *QBLog) Warning(a ...interface{}) {
	log.Println("WARN:", a)
}

func (slf *QBLog) Error(a ...interface{}) {
	log.Println("ERROR:", a)
}
