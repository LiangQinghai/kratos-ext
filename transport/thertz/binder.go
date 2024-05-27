package thertz

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route/param"
)

var thertzBinderInstance binding.Binder = &thertzBinder{defaultBinder: binding.DefaultBinder()}

// binderMid bind params
func binderMid() Handler {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.SetBinder(thertzBinderInstance)
		ctx.Next(c)
	}
}

type thertzBinder struct {
	defaultBinder binding.Binder
}

func (t *thertzBinder) Name() string {
	return "thertz"
}

func (t *thertzBinder) Bind(request *protocol.Request, i interface{}, params param.Params) error {
	return t.defaultBinder.Bind(request, i, params)
}

func (t *thertzBinder) BindAndValidate(request *protocol.Request, i interface{}, params param.Params) error {
	return t.defaultBinder.BindAndValidate(request, i, params)
}

func (t *thertzBinder) BindQuery(request *protocol.Request, i interface{}) error {
	return t.defaultBinder.Bind(request, i, param.Params{})
}

func (t *thertzBinder) BindHeader(request *protocol.Request, i interface{}) error {
	return t.defaultBinder.Bind(request, i, param.Params{})
}

func (t *thertzBinder) BindPath(request *protocol.Request, i interface{}, params param.Params) error {
	return t.defaultBinder.Bind(request, i, params)
}

func (t *thertzBinder) BindForm(request *protocol.Request, i interface{}) error {
	return t.defaultBinder.Bind(request, i, param.Params{})
}

func (t *thertzBinder) BindJSON(request *protocol.Request, i interface{}) error {
	return t.defaultBinder.BindJSON(request, i)
}

func (t *thertzBinder) BindProtobuf(request *protocol.Request, i interface{}) error {
	return t.defaultBinder.BindProtobuf(request, i)

}
