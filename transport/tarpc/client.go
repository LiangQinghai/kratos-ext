package tarpc

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/lesismal/arpc/codec"
	"github.com/lesismal/arpc/extension/micro"
	"github.com/lesismal/arpc/util"
	"net"
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

// WithNodeFilter with select filters
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(o *clientOptions) {
		o.filters = filters
	}
}

// clientOptions is arpc client config
type clientOptions struct {
	endpoint     string
	timeout      time.Duration
	discovery    registry.Discovery
	middleware   []middleware.Middleware
	balancerName string
	filters      []selector.NodeFilter
}

func Dail(ctx context.Context, opts ...ClientOption) (*Client, error) {
	options := clientOptions{
		timeout:      time.Second * 5,
		balancerName: balancerName,
	}
	for _, opt := range opts {
		opt(&options)
	}
	serviceManager := micro.NewServiceManager(
		func(addr string) (net.Conn, error) {
			return net.Dial("tcp", addr)
		})
	if options.discovery != nil {
		watch, err := options.discovery.Watch(ctx, options.endpoint[13:])
		if err != nil {
			return nil, err
		}
		d := &discovery{
			w:                watch,
			serviceNamespace: "defaultServiceNamespace",
			serviceManager:   serviceManager,
			ctx:              ctx,
		}
		go d.watch()
	} else {
		serviceManager.AddServiceNodes(fmt.Sprintf("defaultServiceNamespace/%s/%s", options.endpoint, options.endpoint), "10")
	}
	return &Client{
		opts:           &options,
		serviceManager: serviceManager,
	}, nil
}

type Client struct {
	opts           *clientOptions
	serviceManager micro.ServiceManager
}

func (c *Client) Call(ctx context.Context, method string, req any, resp any) error {

	ctx = transport.NewClientContext(ctx, &Transport{
		operation: method,
		reqHeader: headerCarrier{},
	})

	var h middleware.Handler = func(ctx context.Context, req interface{}) (interface{}, error) {
		reqMsg := c.newMessage(ctx, req)
		var replyMsg MessageWrapper
		var err error
		ac, err := c.serviceManager.ClientBy(c.opts.endpoint)
		if err != nil {
			return nil, err
		}
		if c.opts.timeout > 0 {
			err = ac.Call(
				method,
				reqMsg,
				&replyMsg,
				c.opts.timeout,
			)
		} else {
			err = ac.CallWith(
				ctx,
				method,
				reqMsg,
				&replyMsg,
			)
		}
		if err != nil {
			return nil, err
		}
		if replyMsg.Err != nil {
			return nil, replyMsg.Err
		}
		if replyMsg.Data == nil {
			return nil, nil
		}
		err = util.BytesToValue(codec.DefaultCodec, replyMsg.Data, resp)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	if len(c.opts.middleware) > 0 {
		h = middleware.Chain(c.opts.middleware...)(h)
	}
	var p selector.Peer
	ctx = selector.NewPeerContext(ctx, &p)
	_, err := h(ctx, req)
	return err
}

func (c *Client) newMessage(ctx context.Context, data any) *MessageWrapper {
	if err, ok := data.(error); ok {
		return c.errorEncode(ctx, err)
	}
	headers := c.parseHeader(ctx)
	return &MessageWrapper{
		Headers: headers,
		Data:    util.ValueToBytes(codec.DefaultCodec, data),
	}
}

func (c *Client) parseHeader(ctx context.Context) map[string][]string {
	if tr, ok := transport.FromClientContext(ctx); ok {
		header := tr.RequestHeader()
		keys := header.Keys()
		keyVals := make(map[string][]string)
		for _, k := range keys {
			keyVals[k] = header.Values(k)
		}
		return keyVals
	}
	return nil
}

func (c *Client) errorEncode(ctx context.Context, err error) *MessageWrapper {
	se := errors.FromError(err)
	return &MessageWrapper{
		Headers: c.parseHeader(ctx),
		Err:     se,
	}
}
