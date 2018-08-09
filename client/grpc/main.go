package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/chris-rock/gyrpsy/api/pingpong"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	serverAddr := "localhost:5001"

	creds, err := credentials.NewClientTLSFromFile("./cert/cert_localhost_5001.pem", "")
	if err != nil {
		panic(err)
	}
	creds.OverrideServerName("localhost:5001")

	// load certifice
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	client := pingpong.NewPingPongClient(conn)
	version, err := client.Ping(context.Background(), &pingpong.PingRequest{Sender: "John"})
	if err != nil {
		panic(err)
	}
	log.Info(version)
}
