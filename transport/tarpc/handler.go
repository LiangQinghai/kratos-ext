package tarpc

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/lesismal/arpc"
)

type HandlerFunc = arpc.HandlerFunc

func RecoveryHandler() HandlerFunc {
	return func(ctx *arpc.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic recovery from %v", err)
				mw := defaultErrorEncoder(ctx, err.(error))
				e := ctx.Write(mw)
				if e != nil {
					log.Errorf("recover write err:%v", e)
				}
			}
		}()
		ctx.Next()
	}
}
