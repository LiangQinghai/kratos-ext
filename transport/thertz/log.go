package thertz

import (
	"context"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/go-kratos/kratos/v2/log"
	"io"
)

func init() {
	hlog.SetLogger(newThertzLog())
}

func SetLogger(logger log.Logger) {
	hlog.SetLogger(&thertzLog{
		log: log.NewHelper(logger),
	})
}

func newThertzLog() hlog.FullLogger {
	return &thertzLog{
		log: log.NewHelper(log.DefaultLogger),
	}
}

type thertzLog struct {
	log *log.Helper
}

func (t *thertzLog) Trace(v ...interface{}) {
	t.log.Debug(v...)
}

func (t *thertzLog) Debug(v ...interface{}) {
	t.log.Debug(v...)
}

func (t *thertzLog) Info(v ...interface{}) {
	t.log.Info(v...)
}

func (t *thertzLog) Notice(v ...interface{}) {
	t.log.Info(v...)
}

func (t *thertzLog) Warn(v ...interface{}) {
	t.log.Warn(v...)
}

func (t *thertzLog) Error(v ...interface{}) {
	t.log.Error(v...)
}

func (t *thertzLog) Fatal(v ...interface{}) {
	t.log.Fatal(v...)
}

func (t *thertzLog) Tracef(format string, v ...interface{}) {
	t.log.Debugf(format, v...)
}

func (t *thertzLog) Debugf(format string, v ...interface{}) {
	t.log.Debugf(format, v...)
}

func (t *thertzLog) Infof(format string, v ...interface{}) {
	t.log.Infof(format, v...)
}

func (t *thertzLog) Noticef(format string, v ...interface{}) {
	t.log.Infof(format, v...)
}

func (t *thertzLog) Warnf(format string, v ...interface{}) {
	t.log.Warnf(format, v...)
}

func (t *thertzLog) Errorf(format string, v ...interface{}) {
	t.log.Errorf(format, v...)
}

func (t *thertzLog) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func (t *thertzLog) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	t.log.WithContext(ctx).Debugf(format, v...)
}

func (t *thertzLog) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	t.log.WithContext(ctx).Debugf(format, v...)
}

func (t *thertzLog) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	t.log.WithContext(ctx).Infof(format, v...)
}

func (t *thertzLog) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	t.log.WithContext(ctx).Infof(format, v...)
}

func (t *thertzLog) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	t.log.WithContext(ctx).Warnf(format, v...)
}

func (t *thertzLog) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	t.log.WithContext(ctx).Errorf(format, v...)
}

func (t *thertzLog) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	t.log.WithContext(ctx).Fatalf(format, v...)
}

func (t *thertzLog) SetLevel(_ hlog.Level) {
}

func (t *thertzLog) SetOutput(_ io.Writer) {
}
