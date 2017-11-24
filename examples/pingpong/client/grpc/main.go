package main

import (
	"crypto/x509"

	log "github.com/Sirupsen/logrus"
	pb "github.com/chris-rock/gyrpsy/examples/pingpong/proto"
	"github.com/chris-rock/gyrpsy/server"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	serverAddr := "localhost:5000"

	var opts []grpc.DialOption

	// load generate certificate from server to make this easy here
	// in production, we would lode the file directly
	_, cert := server.GetCertificates(serverAddr)

	// load certifice
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(cert))
	creds := credentials.NewClientTLSFromCert(certPool, serverAddr)
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Error(err)
	}

	defer conn.Close()

	client := pb.NewPingPongClient(conn)
	version, err := client.Ping(context.Background(), &pb.PingRequest{Sender: "John"})
	if err != nil {
		log.Error(err)
	}
	log.Info(version)
}
