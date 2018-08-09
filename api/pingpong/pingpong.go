package pingpong

import "golang.org/x/net/context"
import google_protobuf_empty "github.com/golang/protobuf/ptypes/empty"

//go:generate protoc -I. -I$GOPATH/src -I$PWD/vendor -I$PWD/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:. --grpc-gateway_out=request_context=true,logtostderr=true:. pingpong.proto

// server is used to implement PingPong service
type PingPongServerImpl struct{}

// SayHello implements helloworld.GreeterServer
func (s *PingPongServerImpl) Ping(ctx context.Context, in *PingRequest) (*PongReply, error) {
	return &PongReply{Message: "Hello " + in.GetSender()}, nil
}

func (s *PingPongServerImpl) NoPing(ctx context.Context, in *google_protobuf_empty.Empty) (*PongReply, error) {
	return &PongReply{Message: "HelloPong"}, nil
}
