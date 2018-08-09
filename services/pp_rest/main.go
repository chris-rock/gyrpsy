package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/chris-rock/gyrpsy/api/pingpong"
	server "github.com/chris-rock/gyrpsy/gateway/go/gateway"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var httpsport = 5002
var httpport = 5003
var grpcPort = 5001

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
	HttpAddr := fmt.Sprintf("localhost:%d", httpport)
	HttpsAddr := fmt.Sprintf("localhost:%d", httpsport)
	GrpcAddr := fmt.Sprintf("localhost:%d", grpcPort)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// load certificate from binary
	key, cert := GetCertificates("./cert/key_localhost_5001.pem", "./cert/cert_localhost_5001.pem")

	// set up GRPC client
	_, CertPool := server.LoadCert(cert, key)
	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: GrpcAddr,
		RootCAs:    CertPool,
	})
	opts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	// start mux and attach the grpc client
	mux := runtime.NewServeMux()
	err := pingpong.RegisterPingPongHandlerFromEndpoint(ctx, mux, GrpcAddr, opts)
	if err != nil {
		panic(err)
	}

	// start https server with mux
	rest_key, rest_cert := GetCertificates("./cert/key_localhost_5002.pem", "./cert/cert_localhost_5002.pem")
	RestKeyPair, _ := server.LoadCert(rest_cert, rest_key)

	// start http server in its own go routine
	go func() {
		// start http server with mux
		srv := &http.Server{
			Addr:    HttpAddr,
			Handler: mux,
		}

		lis, err := net.Listen("tcp", HttpAddr)
		if err != nil {
			fmt.Sprintf("%v", err)
		}
		fmt.Printf("http server is listening on %s\n", HttpAddr)
		srv.Serve(lis)
	}()

	// start https server

	srv := &http.Server{
		Addr:    HttpsAddr,
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*RestKeyPair},
		},
	}

	lis, err := net.Listen("tcp", HttpsAddr)
	if err != nil {
		fmt.Sprintf("%v", err)
	}

	tlslis := tls.NewListener(lis, srv.TLSConfig)
	fmt.Printf("https server is listening on %s\n", HttpsAddr)
	srv.Serve(tlslis)
}
