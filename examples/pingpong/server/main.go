package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc"

	"github.com/Sirupsen/logrus"
	pb "github.com/chris-rock/gyrpsy/examples/pingpong/proto"
	"github.com/chris-rock/gyrpsy/server"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
)

// server is used to implement helloworld.GreeterServer.
type pingPongServer struct{}

// SayHello implements helloworld.GreeterServer
func (s *pingPongServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PongReply, error) {
	return &pb.PongReply{Message: "Hello " + in.GetSender()}, nil
}

func (s *pingPongServer) NoPing(ctx context.Context, in *google_protobuf.Empty) (*pb.PongReply, error) {
	return &pb.PongReply{Message: "HelloPong"}, nil
}

var port = 5000

func main() {
	key, cert := server.GetCertificates(fmt.Sprintf("%s:%d", "localhost", port))
	s := server.NewServer(server.Config{
		Hostname: "localhost",
		Port:     port,
		Key:      key,
		Cert:     cert,
	})

	// handle all GRPC services
	grpcHandler := func(opts *server.Options, grpcServer *grpc.Server) (err error) {
		logrus.Info("Register Ping Pong Server")
		pb.RegisterPingPongServer(grpcServer, &pingPongServer{})
		return nil
	}
	s.HandleGRPC(grpcHandler)

	// handle rest gateway services
	restHandler := func(opts *server.Options) (route *http.ServeMux, err error) {
		logrus.Infof("Register gateway to %s", opts.GrpcAddr)
		route = http.NewServeMux()

		pingpongMux := runtime.NewServeMux()
		ctx := context.Background()
		err = pb.RegisterPingPongHandlerFromEndpoint(ctx, pingpongMux, opts.GrpcAddr, opts.Dopts)
		if err != nil {
			logrus.Infof("cannot serve pingpong api: %v\n", err)
			return route, err
		}
		route.Handle("/", pingpongMux)
		return route, nil
	}
	s.Handle("/pingpong/", restHandler)

	// handle optional mux handles
	swaggerHandler := func(opts *server.Options) (mux *http.ServeMux, err error) {
		mux = http.NewServeMux()
		// TODO: read all scopes from pb.Swagger
		mux.HandleFunc("/rest", func(w http.ResponseWriter, req *http.Request) {
			io.Copy(w, strings.NewReader("pingpong"))
		})

		return mux, nil
	}

	s.Handle("/custom/", swaggerHandler)

	// Register reflection service on gRPC server.
	// reflection.Register(s.GRPC)
	if err := s.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
