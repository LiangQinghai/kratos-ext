package tarpc

import (
	"context"
	"crypto/tls"
	"github.com/LiangQinghai/kratos-ext/pkg/endpoint"
	"github.com/LiangQinghai/kratos-ext/pkg/host"
	"github.com/LiangQinghai/kratos-ext/pkg/matcher"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/lesismal/arpc"
	"github.com/lesismal/arpc/codec"
	"github.com/lesismal/arpc/util"
	"google.golang.org/grpc/metadata"
	"net"
	"net/url"
	"time"
)

var (
	_              transport.Server     = (*Server)(nil)
	_              transport.Endpointer = (*Server)(nil)
	NoHandlerError                      = errors.New(404, "handler not found", "no handler found")
)

// ServerOption is an arpc framework option
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

// ErrorEncode with error encode
func ErrorEncode(ene EncodeErrorFunc) ServerOption {
	return func(s *Server) {
		s.ene = ene
		setErrorEncoder(ene)
	}
}

// Recovery error recovery handler
func Recovery(rec HandlerFunc) ServerOption {
	return func(s *Server) {
		s.rec = rec
	}
}

// Middleware mid
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware.Use(m...)
	}
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:    "tcp",
		address:    ":9090",
		middleware: matcher.New(),
		timeout:    3 * time.Second,
		ene:        defaultErrorEncoder,
		rec:        RecoveryHandler(),
	}
	for _, opt := range opts {
		opt(srv)
	}
	arpcServer := arpc.NewServer()
	//recovery
	arpcServer.Handler.Use(srv.rec)
	srv.arpcServer = arpcServer
	return srv
}

type Server struct {
	arpcServer *arpc.Server
	baseCtx    context.Context
	lis        net.Listener
	tlsConf    *tls.Config
	err        error
	network    string
	address    string
	endpoint   *url.URL
	timeout    time.Duration
	middleware matcher.Matcher
	ene        EncodeErrorFunc
	rec        HandlerFunc
}

func (s *Server) Endpoint() (*url.URL, error) {
	err := s.listenAndEndpoint()
	if err != nil {
		return nil, err
	}
	return s.endpoint, nil
}

func (s *Server) Start(ctx context.Context) error {
	err := s.listenAndEndpoint()
	if err != nil {
		return err
	}
	log.Infof("[ARPC] server listening on: %s", s.lis.Addr().String())
	s.baseCtx = ctx
	return s.arpcServer.Serve(s.lis)
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("[ARPC] server stopping")
	return s.arpcServer.Shutdown(ctx)
}

func (s *Server) Handle(m string, handler HandlerFunc) *Server {
	s.arpcServer.Handler.Handle(m, handler)
	return s
}

func (s *Server) Middleware(ctx context.Context, m middleware.Handler) middleware.Handler {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return middleware.Chain(s.middleware.Match(tr.Operation())...)(m)
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, NoHandlerError
	}
}

func (s *Server) DecodeData(data []byte, target any) error {
	return util.BytesToValue(codec.DefaultCodec, data, target)
}

func (s *Server) DecodeRequest(c *Ctx) (context.Context, []byte, error) {
	body := c.Body()
	if body == nil || len(body) == 0 {
		return c, nil, nil
	}
	var mw MessageWrapper
	// decode
	err := util.BytesToValue(codec.DefaultCodec, body, &mw)
	if err != nil {
		return c, nil, err
	}
	// init transport
	ctx := s.initTransport(c, mw.Headers)
	if mw.Err != nil {
		return ctx, nil, mw.Err
	}
	return ctx, mw.Data, nil
}

func (s *Server) EncodeResponse(ctx context.Context, resp any, err error) *MessageWrapper {
	if err != nil {
		mw := s.ene(ctx, err)
		return mw
	}
	if resp == nil {
		return &MessageWrapper{}
	}
	// encode resp
	bytes := util.ValueToBytes(codec.DefaultCodec, resp)
	var md metadata.MD
	if tr, ok := FromArpcTransport(ctx); ok {
		md = metadata.MD(tr.replyHeader)
	}
	// wrap data
	return &MessageWrapper{
		Data:    bytes,
		Headers: md,
	}
}

func (s *Server) Write(c *Ctx, data any) {
	err := c.Write(data)
	if err != nil {
		panic(err)
	}
}

func (s *Server) initTransport(ctx context.Context, reqHeader map[string][]string) context.Context {
	var (
		cancel context.CancelFunc
	)
	if s.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	tr := Transport{
		endpoint:    s.endpoint.String(),
		reqHeader:   mapToHeaderCarrier(reqHeader),
		replyHeader: mapToHeaderCarrier(map[string][]string{}),
	}
	ctx = transport.NewServerContext(ctx, &tr)
	return ctx
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
		addr, err := host.ExtractFromLis(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("http", s.tlsConf != nil), addr)
	}
	return s.err
}
