package cradle

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/atomic"
)

type Server struct {
	listenAddress string
	tlsConfig     *tls.Config
	handler       *mux.Router
	listener      atomic.Value
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}
	if s.tlsConfig != nil {
		listener = tls.NewListener(listener, s.tlsConfig)
	}
	s.listener.Store(listener)
	return http.Serve(listener, s.handler)
}

func (s *Server) Shutdown() {
	if listener, ok := s.listener.Load().(net.Listener); ok && listener != nil {
		_ = listener.Close()
	}
}
