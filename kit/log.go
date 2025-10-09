package kit

import (
	"fmt"

	"github.com/chaos-io/chaos/pkg/logs"
)

type kitLogger struct {
	logs.Logger
}

func Logger() logs.Logger {
	return logs.DefaultLogger()
}

// Log err & level should be the first key
// msg may be first or only after the level
func (l kitLogger) Log(kvs ...interface{}) error {
	length := len(kvs)
	secondMsgField := func(l int, key interface{}) bool { return l >= 4 && fmt.Sprint(key) == "msg" }

	if length >= 2 {
		key := fmt.Sprint(kvs[0])
		switch key {
		case "err":
			l.Errorw("", kvs...)
		case "msg":
			l.Infow(fmt.Sprint(kvs[1]), kvs[2:]...)
		case "level":
			level := fmt.Sprint(kvs[1])
			switch level {
			case "debug":
				if secondMsgField(length, kvs[2]) {
					l.Debugw(fmt.Sprint(kvs[3]), kvs[4:]...)
				} else {
					l.Debugw("", kvs[2:]...)
				}
			default:
				fallthrough
			case "info":
				if secondMsgField(length, kvs[2]) {
					l.Infow(fmt.Sprint(kvs[3]), kvs[4:]...)
				} else {
					l.Infow("", kvs[2:]...)
				}
			case "warn":
				if secondMsgField(length, kvs[2]) {
					l.Warnw(fmt.Sprint(kvs[3]), kvs[4:]...)
				} else {
					l.Warnw("", kvs[2:]...)
				}
			case "error":
				if secondMsgField(length, kvs[2]) {
					l.Errorw(fmt.Sprint(kvs[3]), kvs[4:]...)
				} else {
					l.Errorw("", kvs[2:]...)
				}
			}
		default:
			l.Infow("", kvs...)
		}
	}

	return nil
}
