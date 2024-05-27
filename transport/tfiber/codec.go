package tfiber

import (
	"github.com/LiangQinghai/kratos-ext/pkg/httputil"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/gofiber/fiber/v2"
)

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc = func(ctx *Ctx, v any) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(ctx *Ctx, err error) error

func DefaultErrorEncoder(ctx *Ctx, err error) error {
	se := errors.FromError(err)
	codec, _ := CodecForRequest(ctx, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError)
		return nil
	}
	ctx.Status(int(se.Code))
	_, err = ctx.Write(body)
	return err
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(ctx *Ctx, name string) (encoding.Codec, bool) {
	for _, accept := range ctx.GetReqHeaders()[name] {
		codec := encoding.GetCodec(httputil.ContentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec("json"), false
}

// DefaultResponseEncoder encodes the object to the HTTP response.
func DefaultResponseEncoder(ctx *Ctx, v any) error {
	if v == nil {
		return nil
	}
	codec, _ := CodecForRequest(ctx, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	ctx.Set("Content-Type", httputil.ContentType(codec.Name()))
	_, err = ctx.Write(data)
	if err != nil {
		return err
	}
	return nil
}
