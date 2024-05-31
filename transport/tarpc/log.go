package tarpc

import (
	"github.com/go-kratos/kratos/v2/log"
	arpcLog "github.com/lesismal/arpc/log"
)

func init() {
	arpcLog.SetLogger(new(arpcLogAdapter))
}

type arpcLogAdapter struct {
}

func (a *arpcLogAdapter) SetLevel(lvl int) {
}

func (a *arpcLogAdapter) Debug(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

func (a *arpcLogAdapter) Info(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func (a *arpcLogAdapter) Warn(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

func (a *arpcLogAdapter) Error(format string, v ...interface{}) {
	log.Errorf(format, v...)
}
