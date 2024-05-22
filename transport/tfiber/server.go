package tfiber

import (
	"context"
	"crypto/tls"
	"github.com/LiangQinghai/kratos-advance/transport/tfiber/internal/endpoint"
	"github.com/LiangQinghai/kratos-advance/transport/tfiber/internal/host"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/gofiber/fiber/v2"
	"net"
	"net/url"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
	//_ fiber.Handler        = (*Server)(nil)
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

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: ":0",
		prefork: false,
	}
	for _, opt := range opts {
		opt(srv)
	}
	c := fiber.Config{
		Prefork: srv.prefork,
	}
	srv.App = fiber.New(c)
	return srv
}

type Server struct {
	*fiber.App
	lis      net.Listener
	tlsConf  *tls.Config
	endpoint *url.URL
	err      error
	network  string
	address  string
	prefork  bool
}

func (s *Server) Start(ctx context.Context) error {
	err := s.listenAndEndpoint()
	if err != nil {
		return err
	}
	if s.tlsConf == nil {
		return s.ListenTLS(s.address, "", "")
	}
	return s.Listener(s.lis)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.ShutdownWithContext(ctx)
}

func (s *Server) Endpoint() (*url.URL, error) {
	err := s.listenAndEndpoint()
	if err != nil {
		return nil, err
	}
	return s.endpoint, nil
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