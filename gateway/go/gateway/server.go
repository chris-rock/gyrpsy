package gateway

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	Hostname string
	Port     int
	Key      []byte
	Cert     []byte
}

type Options struct {
	GrpcAddr string
	Dopts    []grpc.DialOption
}

type tlsConfig struct {
	KeyPair  *tls.Certificate
	CertPool *x509.CertPool
}

type Server struct {
	config  *Config
	GRPC    *grpc.Server
	Options *Options
	Mux     *http.ServeMux
	tls     *tlsConfig
}

func LoadCert(cert []byte, key []byte) (*tls.Certificate, *x509.CertPool) {
	var err error

	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		logrus.Fatal(err)
	}
	keyPair := &pair
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(cert))
	if !ok {
		logrus.Fatal(err)
	}
	return keyPair, certPool
}

func NewServer(config Config, opt ...grpc.ServerOption) *Server {
	s := &Server{}
	s.Options = &Options{}

	// load certificate from binary
	s.tls = &tlsConfig{}
	s.tls.KeyPair, s.tls.CertPool = LoadCert(config.Cert, config.Key)

	// set the cofnig for the tcp listener
	s.Options.GrpcAddr = fmt.Sprintf("%s:%d", config.Hostname, config.Port)
	logrus.Infof("Server %s", s.Options.GrpcAddr)

	// set TLS config for the GRPC gateway clients
	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: s.Options.GrpcAddr,
		RootCAs:    s.tls.CertPool,
	})
	s.Options.Dopts = []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	// TODO: merge options with provided ones
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(
			s.tls.CertPool,
			s.Options.GrpcAddr,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_prometheus.StreamServerInterceptor,
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_prometheus.UnaryServerInterceptor,
		)),
	}

	// initialize GRPC server
	s.GRPC = grpc.NewServer(opts...)
	// initialize REST 1.1 handler
	s.Mux = http.NewServeMux()
	return s
}

// handle GRPC returns
func (s *Server) grpcHandlerFunc(grpcServer *grpc.Server, muxHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// partial check of https://github.com/grpc/grpc-go/blob/master/transport/handler_server.go#L50
		if r.ProtoMajor >= 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			muxHandler.ServeHTTP(w, r)
		}
	})
}

func (s *Server) Serve() error {
	logrus.Info("Start server")

	// After all your registrations, make sure all of the Prometheus metrics are initialized.
	grpc_prometheus.Register(s.GRPC)

	// Register Prometheus metrics handler.
	s.Mux.Handle("/metrics", promhttp.Handler())

	// start server
	srv := &http.Server{
		Addr:    s.Options.GrpcAddr,
		Handler: s.grpcHandlerFunc(s.GRPC, s.Mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*s.tls.KeyPair},
			NextProtos:   []string{"h2"},
		},
	}

	lis, err := net.Listen("tcp", s.Options.GrpcAddr)
	if err != nil {
		return errors.Wrapf(err, "listen on %s", s.Options.GrpcAddr)
	}

	tlslis := tls.NewListener(lis, srv.TLSConfig)
	err = srv.Serve(tlslis)
	return errors.Wrap(err, "Serve")
}

// Handle adds a http.ServerMux route to the server. It will serve http 1.1 requests only
func (s *Server) Handle(pattern string, handler func(*Options) (mux *http.ServeMux, err error)) {
	mux, err := handler(s.Options)
	if err != nil {
		logrus.Error(err)
	}

	// strip ending / from subroutes
	l := len(pattern)
	subPattern := pattern
	if string(pattern[l-1]) == "/" {
		subPattern = pattern[0 : l-1]
	}

	logrus.Infof("register route %s", pattern)
	s.Mux.Handle(pattern, http.StripPrefix(subPattern, mux))
}

// HandleGRPC registers a GRPC endpoint to the server
func (s *Server) HandleGRPC(handler func(*Options, *grpc.Server) (err error)) {
	handler(s.Options, s.GRPC)
}
