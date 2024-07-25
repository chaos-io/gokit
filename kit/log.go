package kit

import (
	"fmt"

	"github.com/go-kit/kit/log"

	"github.com/chaos-io/chaos/logs"
)

type kitLogger struct {
	*logs.ZapLogger
}

func Logger() log.Logger {
	return NewKitLogger(logs.Logger())
}

func NewKitLogger(zap *logs.ZapLogger) log.Logger {
	return &kitLogger{zap}
}

// Log err & level should be the first key
// msg may be first or only after the level
func (l kitLogger) Log(keyvals ...interface{}) error {
	length := len(keyvals)
	secondMsgField := func(l int, key interface{}) bool { return l >= 4 && fmt.Sprint(key) == "msg" }

	if length >= 2 {
		key := fmt.Sprint(keyvals[0])
		switch key {
		case "err":
			l.Errorw("", keyvals...)
		case "msg":
			l.Infow(fmt.Sprint(keyvals[1]), keyvals[2:]...)
		case "level":
			level := fmt.Sprint(keyvals[1])
			switch level {
			case "debug":
				if secondMsgField(length, keyvals[2]) {
					l.Debugw(fmt.Sprint(keyvals[3]), keyvals[4:]...)
				} else {
					l.Debugw("", keyvals[2:]...)
				}
			default:
				fallthrough
			case "info":
				if secondMsgField(length, keyvals[2]) {
					l.Infow(fmt.Sprint(keyvals[3]), keyvals[4:]...)
				} else {
					l.Infow("", keyvals[2:]...)
				}
			case "warn":
				if secondMsgField(length, keyvals[2]) {
					l.Warnw(fmt.Sprint(keyvals[3]), keyvals[4:]...)
				} else {
					l.Warnw("", keyvals[2:]...)
				}
			case "error":
				if secondMsgField(length, keyvals[2]) {
					l.Errorw(fmt.Sprint(keyvals[3]), keyvals[4:]...)
				} else {
					l.Errorw("", keyvals[2:]...)
				}
			}
		default:
			l.Infow("", keyvals...)
		}
	}

	return nil
}
