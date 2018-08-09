package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/chris-rock/gyrpsy/api/pingpong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var port = 5001
var portWithoutTls = 6001

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
	creds, err := credentials.NewServerTLSFromFile("./cert/cert_localhost_5001.pem", "./cert/key_localhost_5001.pem")
	if err != nil {
		panic(err)
	}
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
	}

	// start grpc server without tls cert
	go func() {
		grpcServer := grpc.NewServer()
		pingpong.RegisterPingPongServer(grpcServer, &pingpong.PingPongServerImpl{})
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", portWithoutTls))
		if err != nil {
			panic(err)
		}
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	// start grpc server with tls cert
	grpcServer := grpc.NewServer(opts...)
	pingpong.RegisterPingPongServer(grpcServer, &pingpong.PingPongServerImpl{})

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic(err)
	}

	err = grpcServer.Serve(lis)
	if err != nil {
		panic(err)
	}
}
