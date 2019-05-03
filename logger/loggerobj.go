package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	INFO  int = iota
	ERROR int = 1
	FATAL int = 2
)

type Error struct {
	LogType  string `json:"log_type"`
	FileName string `json:"file_name"`
	ErrFile  string `json:"error_file"`
	StrNum   int    `json:"error_row"`
	Message  string `json:"message"`
	FuncName string `json:"func_name"`
}

func NewError(log_file string, msg string, log_type int) *Error {
	e := Error{
		FileName: log_file,
		Message:  msg,
	}
	switch log_type {
	case INFO:
		e.LogType = "INFO"
	case ERROR:
		e.LogType = "ERROR"
	case FATAL:
		e.LogType = "FATAL"
	default:
		e.LogType = "FATAL"
	}
	pc, errFile, strNum, ok := runtime.Caller(1)
	e.ErrFile = errFile
	e.StrNum = strNum
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		e.FuncName = details.Name()
	}
	return &e
}

type Log struct {
	errMsg chan *Error
	files  sync.Map
}

func NewLogger() *Log {
	return &Log{
		errMsg: make(chan *Error, 100),
	}
}

func (l *Log) Println(event *Error) {
	select {
	case l.errMsg <- event:
	}
}

func (l *Log) Subscriber(ctx context.Context) {
	defer func() {
		close(l.errMsg)
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				l.files.Range(func(k, v interface{}) bool {
					v.(*os.File).Close()
					return true
				})
				return
			}
		}
		l.files.Range(func(k, v interface{}) bool {
			v.(*os.File).Close()
			return true
		})
		return
	}()
Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case msg, ok := <-l.errMsg:
			if ok {
				if file, isLoad := l.files.LoadOrStore((*msg).FileName, struct{}{}); !isLoad {
					f, err := os.OpenFile(fmt.Sprintf("./logs/%s_%s.log", (*msg).FileName, time.Now().Format("2006-01-02")),
						os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
					if err != nil {
						panic(err)
					}
					//defer to close when you're done with it, not because you think it's idiomatic!
					l.files.Store(msg.FileName, f)
					log.SetOutput(f)
				} else {
					log.SetOutput(file.(*os.File))
				}
				//(*msg).TimeString = (*msg).Time.Format("2006/01/02 15:04:05")
				if jsonData, err := json.Marshal(*msg); err == nil {
					if (*msg).LogType != "FATAL" {
						log.Output(2, string(jsonData))
					} else {
						log.Fatal(string(jsonData))
					}
				}
			}
		}
	}
	return
}
