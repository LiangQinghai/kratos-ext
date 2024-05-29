package tarpc

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/grpc/metadata"
)

type MessageWrapper struct {
	Data    []byte              `json:"data"`
	Headers map[string][]string `json:"headers"`
	Err     *errors.Error       `json:"error"`
}

var defaultErrorEncoder EncodeErrorFunc = DefaultErrorEncoder

// setErrorEncoder reset error handler
func setErrorEncoder(encoder EncodeErrorFunc) {
	defaultErrorEncoder = encoder
}

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(ctx context.Context, err error) *MessageWrapper

func DefaultErrorEncoder(ctx context.Context, err error) *MessageWrapper {
	se := errors.FromError(err)
	var md metadata.MD
	if tr, ok := FromArpcTransport(ctx); ok {
		md = metadata.MD(tr.replyHeader)
	}
	mw := &MessageWrapper{
		Headers: md,
		Err:     se,
	}
	return mw
}
