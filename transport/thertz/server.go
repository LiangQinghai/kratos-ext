package thertz

import (
	"context"
	"crypto/tls"
	"github.com/LiangQinghai/kratos-ext/pkg/endpoint"
	"github.com/LiangQinghai/kratos-ext/pkg/host"
	"github.com/LiangQinghai/kratos-ext/pkg/matcher"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"net"
	"net/url"
	"time"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

// ServerOption is a fiber framework option
type ServerOption func(*Server)

// Network set network
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with address
func Address(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

// Endpoint with endpoint
func Endpoint(endpoint *url.URL) ServerOption {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

// Middleware with service middleware option.
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(o *Server) {
		o.middleware.Use(m...)
	}
}

// ResponseEncoder with response encoder.
func ResponseEncoder(en EncodeResponseFunc) ServerOption {
	return func(o *Server) {
		o.enc = en
	}
}

// ErrorEncoder with error encoder.
func ErrorEncoder(en EncodeErrorFunc) ServerOption {
	return func(o *Server) {
		o.ene = en
	}
}

// AppName set server app name
func AppName(name string) ServerOption {
	return func(s *Server) {
		s.appName = name
	}
}

// RawMiddleware fiber mid
func RawMiddleware(h ...Handler) ServerOption {
	return func(s *Server) {
		s.rawMid = append(s.rawMid, h...)
	}
}

// NoRouteHandler 404 handler
func NoRouteHandler(h Handler) ServerOption {
	return func(s *Server) {
		s.notFoundHandler = h
	}
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:         "tcp",
		address:         ":8080",
		middleware:      matcher.New(),
		enc:             DefaultResponseEncoder,
		ene:             DefaultErrorEncoder,
		notFoundHandler: Default404Handler,
		timeout:         3 * time.Second,
	}
	for _, opt := range opts {
		opt(srv)
	}
	hOpts := make([]config.Option, 0)
	hOpts = append(hOpts, server.WithNetwork(srv.network))
	hOpts = append(hOpts, server.WithHostPorts(srv.address))
	if srv.tlsConf != nil {
		hOpts = append(hOpts, server.WithTLS(srv.tlsConf))
	}
	hertz := server.New(hOpts...)
	srv.app = hertz
	// error handler
	srv.app.Use(recovery.Recovery(recovery.WithRecoveryHandler(srv.ene)))
	// 404
	srv.app.NoRoute(srv.notFoundHandler)
	return srv
}

type Server struct {
	appName         string
	app             *server.Hertz
	lis             net.Listener
	tlsConf         *tls.Config
	endpoint        *url.URL
	err             error
	network         string
	address         string
	timeout         time.Duration
	middleware      matcher.Matcher
	rawMid          []Handler
	notFoundHandler Handler
	enc             EncodeResponseFunc
	ene             EncodeErrorFunc
}

func (s *Server) Start(ctx context.Context) error {
	err := s.initEndpoint()
	if err != nil {
		return err
	}
	return s.app.Run()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.app.Shutdown(ctx)
}

func (s *Server) Endpoint() (*url.URL, error) {
	err := s.initEndpoint()
	if err != nil {
		return nil, err
	}
	return s.endpoint, nil
}

func (s *Server) Middleware(m middleware.Handler, ctx context.Context, path string) middleware.Handler {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return middleware.Chain(s.middleware.Match(tr.Operation())...)(m)
	}
	return middleware.Chain(s.middleware.Match(path)...)(m)
}

func (s *Server) Router() route.IRoutes {
	r := s.app.Use(binderMid(), s.transportMid())
	if s.rawMid != nil && len(s.rawMid) > 0 {
		for _, h := range s.rawMid {
			r = s.app.Use(h)
		}
	}
	return r
}

// Write response data encode
func (s *Server) Write(ctx *ReqCtx, v any) {
	s.enc(ctx, v)
}

func (s *Server) initEndpoint() error {
	if s.endpoint == nil {
		addr, err := host.Extract(s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("http", s.tlsConf != nil), addr)
	}
	return s.err
}

func (s *Server) transportMid() Handler {
	return func(c context.Context, ctx *app.RequestContext) {
		var cancel context.CancelFunc
		if s.timeout > 0 {
			c, cancel = context.WithTimeout(c, s.timeout)
		} else {
			c, cancel = context.WithCancel(c)
		}
		defer cancel()
		tr := Transport{
			endpoint:    s.endpoint.String(),
			reqHeader:   &requestHeaderCarrier{RequestHeader: &ctx.Request.Header},
			replyHeader: &responseHeaderCarrier{ResponseHeader: &ctx.Response.Header},
			request:     &ctx.Request,
		}
		c = transport.NewServerContext(c, &tr)
		ctx.Next(c)
	}
}
