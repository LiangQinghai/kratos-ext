package thertz

import (
	"context"
	"github.com/LiangQinghai/kratos-ext/transport/thertz/internal/schema"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route/param"
)

var thertzBinderInstance = newThertzBinder()

// binderMid bind params
func binderMid() Handler {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.SetBinder(thertzBinderInstance)
		ctx.Next(c)
	}
}

func newThertzBinder() binding.Binder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag("json")
	return &thertzBinder{
		defaultBinder: binding.DefaultBinder(),
		decoder:       decoder,
	}
}

type thertzBinder struct {
	defaultBinder binding.Binder
	decoder       *schema.Decoder
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
	return t.decoder.Decode(i, t.queryToMap(request))
}

func (t *thertzBinder) BindHeader(request *protocol.Request, i interface{}) error {
	return t.decoder.Decode(i, t.headerToMap(request))
}

func (t *thertzBinder) BindPath(_ *protocol.Request, i interface{}, params param.Params) error {
	res := t.paramsToMap(params)
	return t.decoder.Decode(i, res)
}

func (t *thertzBinder) BindForm(request *protocol.Request, i interface{}) error {
	return t.decoder.Decode(i, t.formToMap(request))
}

func (t *thertzBinder) BindJSON(request *protocol.Request, i interface{}) error {
	return t.defaultBinder.BindJSON(request, i)
}

func (t *thertzBinder) BindProtobuf(request *protocol.Request, i interface{}) error {
	return t.defaultBinder.BindProtobuf(request, i)
}

func (t *thertzBinder) paramsToMap(params param.Params) (res map[string][]string) {
	if params == nil {
		return nil
	}
	res = make(map[string][]string)
	for _, p := range params {
		res[p.Key] = []string{p.Value}
	}
	return
}

func (t *thertzBinder) formToMap(req *protocol.Request) (res map[string][]string) {
	res = make(map[string][]string)
	req.PostArgs().VisitAll(func(key, value []byte) {
		res[string(key)] = []string{string(value)}
	})
	form, err := req.MultipartForm()
	if err != nil || form == nil || form.Value == nil {
		return
	}
	for k, v := range form.Value {
		res[k] = v
	}
	return
}

func (t *thertzBinder) queryToMap(req *protocol.Request) (res map[string][]string) {
	res = make(map[string][]string)
	req.URI().QueryArgs().VisitAll(func(key, value []byte) {
		res[string(key)] = []string{string(value)}
	})
	return
}

func (t *thertzBinder) headerToMap(req *protocol.Request) (res map[string][]string) {
	res = make(map[string][]string)
	req.Header.VisitAll(func(key, value []byte) {
		res[string(key)] = []string{string(value)}
	})
	return
}
