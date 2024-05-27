package tfiber

import (
	"context"
	"crypto/tls"
	"github.com/LiangQinghai/kratos-ext/pkg/endpoint"
	"github.com/LiangQinghai/kratos-ext/pkg/host"
	"github.com/LiangQinghai/kratos-ext/pkg/matcher"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/gofiber/fiber/v2"
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

// Prefork with prefork
func Prefork(prefork bool) ServerOption {
	return func(s *Server) {
		s.fiberConfig.Prefork = prefork
	}
}

// Middleware with service middleware option.
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware.Use(m...)
	}
}

// ResponseEncoder with response encoder.
func ResponseEncoder(en EncodeResponseFunc) ServerOption {
	return func(s *Server) {
		s.enc = en
	}
}

// ErrorEncoder with error encoder.
func ErrorEncoder(en EncodeErrorFunc) ServerOption {
	return func(s *Server) {
		s.fiberConfig.ErrorHandler = en
	}
}

// AppName set server app name
func AppName(name string) ServerOption {
	return func(s *Server) {
		s.fiberConfig.AppName = name
	}
}

// RawMiddleware fiber mid
func RawMiddleware(h ...fiber.Handler) ServerOption {
	return func(s *Server) {
		s.rawMid = append(s.rawMid, h...)
	}
}

// FiberConfig fiber config
func FiberConfig(c *fiber.Config) ServerOption {
	return func(s *Server) {

	}
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:    "tcp",
		address:    ":0",
		middleware: matcher.New(),
		enc:        DefaultResponseEncoder,
		timeout:    3 * time.Second,
		fiberConfig: &fiber.Config{
			ErrorHandler: DefaultErrorEncoder,
		},
	}
	for _, opt := range opts {
		opt(srv)
	}
	srv.app = fiber.New(*srv.fiberConfig)
	return srv
}

type Server struct {
	app         *fiber.App
	tlsConf     *tls.Config
	endpoint    *url.URL
	err         error
	network     string
	address     string
	timeout     time.Duration
	middleware  matcher.Matcher
	rawMid      []fiber.Handler
	enc         EncodeResponseFunc
	fiberConfig *fiber.Config
}

func (s *Server) Start(_ context.Context) error {
	err := s.initEndpoint()
	if err != nil {
		return err
	}
	if s.tlsConf != nil {
		return s.app.ListenTLSWithCertificate(s.address, s.tlsConf.Certificates[0])
	}
	return s.app.Listen(s.address)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) Endpoint() (*url.URL, error) {
	err := s.initEndpoint()
	if err != nil {
		return nil, err
	}
	return s.endpoint, nil
}

// Middleware mid handler
// m: middleware.Handler kratos middleware
// ctx: context.Context
// path: router path
// returns: middleware.Handler
func (s *Server) Middleware(m middleware.Handler, ctx context.Context, path string) middleware.Handler {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return middleware.Chain(s.middleware.Match(tr.Operation())...)(m)
	}
	return middleware.Chain(s.middleware.Match(path)...)(m)
}

// Group router group, it will use Router function
// returns: fiber.Router
func (s *Server) Group(prefix string, h ...Handler) fiber.Router {
	return s.Router().Group(prefix, h...)
}

// Router new router, use transportMid function and rawMid
// returns: fiber.Router
func (s *Server) Router() fiber.Router {
	r := s.app.Use(s.transportMid())
	if s.rawMid != nil && len(s.rawMid) > 0 {
		for _, h := range s.rawMid {
			r = s.app.Use(h)
		}
	}
	return r
}

// Write response data encode
// returns error
func (s *Server) Write(ctx *Ctx, v any) error {
	return s.enc(ctx, v)
}

// Static fiber static file server handler
func (s *Server) Static(prefix, root string, config ...fiber.Static) fiber.Router {
	return s.app.Static(prefix, root, config...)
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

func (s *Server) transportMid() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if s.timeout > 0 {
			ctx, cancel = context.WithTimeout(c.UserContext(), s.timeout)
		} else {
			ctx, cancel = context.WithCancel(c.UserContext())
		}
		defer cancel()
		tr := Transport{
			endpoint:    s.endpoint.String(),
			reqHeader:   c.GetReqHeaders(),
			replyHeader: c.GetRespHeaders(),
			reqCtx:      c,
		}
		c.SetUserContext(transport.NewServerContext(ctx, &tr))
		return c.Next()
	}
}
