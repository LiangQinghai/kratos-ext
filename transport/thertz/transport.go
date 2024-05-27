package thertz

import (
	"context"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	KindFiber transport.Kind = "hertz"
	// SupportPackageIsVersion1 These constants should not be referenced from any other code.
	SupportPackageIsVersion1 = true
)

type Transport struct {
	endpoint    string
	operation   string
	reqHeader   *requestHeaderCarrier
	replyHeader *responseHeaderCarrier
	request     *protocol.Request
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
type requestHeaderCarrier struct {
	*protocol.RequestHeader
}

func (h *requestHeaderCarrier) Keys() []string {
	keys := make([]string, 0)
	h.VisitAll(func(key, _ []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func (h *requestHeaderCarrier) Values(key string) []string {
	return h.GetAll(key)
}

type responseHeaderCarrier struct {
	*protocol.ResponseHeader
}

func (h *responseHeaderCarrier) Keys() []string {
	keys := make([]string, 0)
	h.VisitAll(func(key, _ []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

func (h *responseHeaderCarrier) Values(key string) []string {
	return h.GetAll(key)
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
func RequestFromServerContext(ctx context.Context) (*protocol.Request, bool) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			return tr.request, true
		}
	}
	return nil, false
}
