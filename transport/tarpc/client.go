package tarpc

import (
	"context"
	"crypto/tls"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/lesismal/arpc"
	"time"
)

const (
	balancerName = "selector"
)

func init() {
	if selector.GlobalSelector() == nil {
		selector.SetGlobalSelector(wrr.NewBuilder())
	}
}

// ClientOption is arpc client option.
type ClientOption func(o *clientOptions)

// WithEndpoint with client endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithTimeout with client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithMiddleware with client middleware.
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(o *clientOptions) {
		o.middleware = m
	}
}

// WithDiscovery with client discovery.
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// WithTLSConfig with TLS config.
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsConf = c
	}
}

// WithNodeFilter with select filters
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(o *clientOptions) {
		o.filters = filters
	}
}

// clientOptions is arpc client config
type clientOptions struct {
	endpoint     string
	tlsConf      *tls.Config
	timeout      time.Duration
	discovery    registry.Discovery
	middleware   []middleware.Middleware
	balancerName string
	filters      []selector.NodeFilter
}

func Dail(ctx context.Context, opts ...ClientOption) (*arpc.Client, error) {
	options := clientOptions{
		timeout:      time.Second * 5,
		balancerName: balancerName,
	}
	for _, opt := range opts {
		opt(&options)
	}
	return nil, nil
}
