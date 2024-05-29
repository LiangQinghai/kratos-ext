package tarpc

import (
	"context"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	KindArpc transport.Kind = "arpc"
	// SupportPackageIsVersion1 These constants should not be referenced from any other code.
	SupportPackageIsVersion1 = true
)

type Transport struct {
	endpoint    string
	operation   string
	reqHeader   headerCarrier
	replyHeader headerCarrier
	nodeFilters []selector.NodeFilter
}

// Kind returns the transport kind.
func (tr *Transport) Kind() transport.Kind {
	return KindArpc
}

// Endpoint returns the transport endpoint.
func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

// Operation returns the transport operation.
func (tr *Transport) Operation() string {
	return tr.operation
}

// RequestHeader returns the request header.
func (tr *Transport) RequestHeader() transport.Header {
	return tr.reqHeader
}

// ReplyHeader returns the reply header.
func (tr *Transport) ReplyHeader() transport.Header {
	return tr.replyHeader
}

// NodeFilters returns the client select filters.
func (tr *Transport) NodeFilters() []selector.NodeFilter {
	return tr.nodeFilters
}

type headerCarrier metadata.MD

// Get returns the value associated with the passed key.
func (mc headerCarrier) Get(key string) string {
	vals := metadata.MD(mc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Set stores the key-value pair.
func (mc headerCarrier) Set(key string, value string) {
	metadata.MD(mc).Set(key, value)
}

// Add append value to key-values pair.
func (mc headerCarrier) Add(key string, value string) {
	metadata.MD(mc).Append(key, value)
}

// Keys lists the keys stored in this carrier.
func (mc headerCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range metadata.MD(mc) {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a slice of values associated with the passed key.
func (mc headerCarrier) Values(key string) []string {
	return metadata.MD(mc).Get(key)
}

// mapToHeaderCarrier map converter to header carrier
func mapToHeaderCarrier(m map[string][]string) headerCarrier {
	md := make(headerCarrier, len(m))
	for k, val := range m {
		key := strings.ToLower(k)
		md[key] = val
	}
	return md
}

// FromArpcTransport get arpc Transport from context
func FromArpcTransport(ctx context.Context) (*Transport, bool) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.(*Transport), true
	}
	return nil, false
}

// SetOperation sets the transport operation.
func SetOperation(ctx context.Context, op string) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			tr.operation = op
		}
	}
}
