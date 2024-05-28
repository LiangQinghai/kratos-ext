package tarpc

import (
	"context"
	"crypto/tls"
	"github.com/LiangQinghai/kratos-ext/pkg/endpoint"
	"github.com/LiangQinghai/kratos-ext/pkg/host"
	"github.com/LiangQinghai/kratos-ext/pkg/matcher"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/lesismal/arpc"
	"net"
	"net/url"
	"time"
)

type Server struct {
	*arpc.Server
	baseCtx    context.Context
	lis        net.Listener
	tlsConf    *tls.Config
	err        error
	network    string
	address    string
	endpoint   *url.URL
	timeout    time.Duration
	middleware matcher.Matcher
}

func (s *Server) Start(ctx context.Context) error {
	err := s.listenAndEndpoint()
	if err != nil {
		return err
	}
	log.Infof("[ARPC] server listening on: %s", s.lis.Addr().String())
	s.baseCtx = ctx
	return s.Serve(s.lis)
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("[ARPC] server stopping")
	return s.Shutdown(ctx)
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
