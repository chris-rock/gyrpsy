package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/chris-rock/gyrpsy/api/pingpong"
	server "github.com/chris-rock/gyrpsy/gateway/go/gateway"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var port = 5000

func GetCertificates(keyFilename string, certFilename string) (ko []byte, co []byte) {
	// load key from disk
	var err error
	ko, err = ioutil.ReadFile(keyFilename)
	if err != nil {
		log.Fatalf("failed to open %s for reading: %s", keyFilename, err)
		panic("could not load tls key")
	}
	co, err = ioutil.ReadFile(certFilename)
	if err != nil {
		log.Fatalf("failed to open %s for reading: %s", certFilename, err)
		panic("could not load tls cert")
	}
	return ko, co
}

func main() {
	key, cert := GetCertificates("./cert/key_localhost_5000.pem", "./cert/cert_localhost_5000.pem")
	s := server.NewServer(server.Config{
		Hostname: "localhost",
		Port:     port,
		Key:      key,
		Cert:     cert,
	})

	// handle all GRPC services
	grpcHandler := func(opts *server.Options, grpcServer *grpc.Server) (err error) {
		logrus.Info("Register Ping Pong Server")
		pingpong.RegisterPingPongServer(grpcServer, &pingpong.PingPongServerImpl{})
		return nil
	}
	s.HandleGRPC(grpcHandler)

	// handle rest gateway services
	restHandler := func(opts *server.Options) (route *http.ServeMux, err error) {
		logrus.Infof("Register gateway to %s", opts.GrpcAddr)
		route = http.NewServeMux()

		pingpongMux := runtime.NewServeMux()
		ctx := context.Background()
		err = pingpong.RegisterPingPongHandlerFromEndpoint(ctx, pingpongMux, opts.GrpcAddr, opts.Dopts)
		if err != nil {
			logrus.Infof("cannot serve pingpong api: %v\n", err)
			return route, err
		}
		route.Handle("/", pingpongMux)
		return route, nil
	}
	s.Handle("/pingpong/", restHandler)

	// handle optional mux handles
	custsomRoute := func(opts *server.Options) (mux *http.ServeMux, err error) {
		mux = http.NewServeMux()
		// TODO: read all scopes from pb.Swagger
		mux.HandleFunc("/rest", func(w http.ResponseWriter, req *http.Request) {
			io.Copy(w, strings.NewReader("pingpong"))
		})

		return mux, nil
	}

	s.Handle("/custom/", custsomRoute)

	// Register reflection service on gRPC server.
	// reflection.Register(s.GRPC)
	if err := s.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
