package thertz

import (
	"context"
	"github.com/LiangQinghai/kratos-ext/pkg/httputil"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/errors"
	"net/http"
)

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc = func(ctx *app.RequestContext, v any)

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte)

func DefaultErrorEncoder(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
	se := errors.FromError(err.(error))
	codec, _ := CodecForRequest(ctx, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(int(se.Code))
	_, _ = ctx.Write(body)
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(ctx *app.RequestContext, name string) (encoding.Codec, bool) {
	for _, accept := range ctx.Request.Header.GetAll(name) {
		codec := encoding.GetCodec(httputil.ContentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec("json"), false
}

// DefaultResponseEncoder encodes the object to the HTTP response.
func DefaultResponseEncoder(ctx *app.RequestContext, v any) {
	if v == nil {
		return
	}
	codec, _ := CodecForRequest(ctx, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		panic(err)
	}
	ctx.Set("Content-Type", httputil.ContentType(codec.Name()))
	_, err = ctx.Write(data)
	if err != nil {
		panic(err)
	}
}
