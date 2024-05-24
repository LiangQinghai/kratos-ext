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
	"net"
	"net/url"
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

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:    "tcp",
		address:    ":0",
		prefork:    false,
		middleware: matcher.New(),
		enc:        DefaultResponseEncoder,
		ene:        DefaultErrorEncoder,
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
	lis        net.Listener
	tlsConf    *tls.Config
	endpoint   *url.URL
	err        error
	network    string
	address    string
	prefork    bool
	middleware matcher.Matcher
	enc        EncodeResponseFunc
	ene        EncodeErrorFunc
}

func (s *Server) Start(ctx context.Context) error {
	err := s.listenAndEndpoint()
	if err != nil {
		return err
	}
	if s.tlsConf != nil {
		return s.app.ListenTLS(s.address, "", "")
	}
	return s.app.Listener(s.lis)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

func (s *Server) Endpoint() (*url.URL, error) {
	err := s.listenAndEndpoint()
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
	return s.app.Group(prefix)
}

// Write response data encode
func (s *Server) Write(ctx *Ctx, v any) error {
	return s.enc(ctx, v)
}

func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		addr, err := host.Extract(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("http", s.tlsConf != nil), addr)
	}
	return s.err
}
