package tfiber

import (
	"context"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	KindFiber transport.Kind = "fiber"
	// SupportPackageIsVersion1 These constants should not be referenced from any other code.
	SupportPackageIsVersion1 = true
)

type Transport struct {
	endpoint    string
	operation   string
	reqHeader   headerCarrier
	replyHeader headerCarrier
	reqCtx      *Ctx
}

func (t *Transport) Kind() transport.Kind {
	return KindFiber
}

func (t *Transport) Endpoint() string {
	return t.endpoint
}

func (t *Transport) Operation() string {
	return t.operation
}

func (t *Transport) RequestHeader() transport.Header {
	return t.reqHeader
}

func (t *Transport) ReplyHeader() transport.Header {
	return t.replyHeader
}

// header
type headerCarrier map[string][]string

func (h headerCarrier) Get(key string) string {
	if v, ok := h[key]; ok {
		return v[0]
	}
	return ""
}

func (h headerCarrier) Set(key string, value string) {
	h[key] = []string{value}
}

func (h headerCarrier) Add(key string, value string) {
	if _, ok := h[key]; ok {
		h[key] = append(h[key], value)
	} else {
		h[key] = []string{value}
	}
}

func (h headerCarrier) Keys() []string {
	keys := make([]string, len(h))
	for k, _ := range h {
		keys = append(keys, k)
	}
	return keys
}

func (h headerCarrier) Values(key string) []string {
	return h[key]
}

// SetOperation sets the transport operation.
func SetOperation(ctx context.Context, op string) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			tr.operation = op
		}
	}
}

// RequestFromServerContext returns request from context.
func RequestFromServerContext(ctx context.Context) (*Ctx, bool) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			return tr.reqCtx, true
		}
	}
	return nil, false
}
