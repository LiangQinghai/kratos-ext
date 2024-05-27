package tfiber

import (
	"context"
	"crypto/tls"
	"github.com/LiangQinghai/kratos-ext/transport/tfiber/internal/endpoint"
	"github.com/LiangQinghai/kratos-ext/transport/tfiber/internal/host"
	"github.com/LiangQinghai/kratos-ext/transport/tfiber/internal/matcher"
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
		s.prefork = prefork
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
func RawMiddleware(h ...fiber.Handler) ServerOption {
	return func(s *Server) {
		s.rawMid = append(s.rawMid, h...)
	}
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:    "tcp",
		address:    ":0",
		prefork:    false,
		middleware: matcher.New(),
		enc:        DefaultResponseEncoder,
		ene:        DefaultErrorEncoder,
		timeout:    3 * time.Second,
	}
	for _, opt := range opts {
		opt(srv)
	}
	c := fiber.Config{
		Prefork:      srv.prefork,
		ErrorHandler: srv.ene,
	}
	srv.app = fiber.New(c)
	return srv
}

type Server struct {
	appName    string
	app        *fiber.App
	tlsConf    *tls.Config
	endpoint   *url.URL
	err        error
	network    string
	address    string
	timeout    time.Duration
	prefork    bool
	middleware matcher.Matcher
	rawMid     []fiber.Handler
	enc        EncodeResponseFunc
	ene        EncodeErrorFunc
}

func (s *Server) Start(ctx context.Context) error {
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

func (s *Server) Middleware(m middleware.Handler, ctx context.Context, path string) middleware.Handler {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return middleware.Chain(s.middleware.Match(tr.Operation())...)(m)
	}
	return middleware.Chain(s.middleware.Match(path)...)(m)
}

func (s *Server) Group(prefix string) fiber.Router {
	r := s.app.Use(s.transportMid())
	if s.rawMid != nil && len(s.rawMid) > 0 {
		for _, h := range s.rawMid {
			r = s.app.Use(h)
		}
	}
	return r.Group(prefix)
}

// Write response data encode
func (s *Server) Write(ctx *Ctx, v any) error {
	return s.enc(ctx, v)
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
