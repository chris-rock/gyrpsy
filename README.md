# Gyrpsy

This is an experiment to see how easy we can auto-generate REST APIs. The implementation uses [Golang](https://golang.org/) and [GRPC](https://grpc.io/) as fundamental concepts.

## Design

- opinionated way to write REST services
- use fast golang server
- all REST endpoints are auto-generated except for multipart uploads
- focus on business functionality

The idea is that we use an abstract [protobuf]() description to generate GRPC and REST endpoints. The sample ping pong service looks as following:

```
service PingPong {
  rpc Ping (PingRequest) returns (PongReply) {
    option (google.api.http) = {
      post: "/ping"
      body: "*"
    };
  };

  rpc NoPing(google.protobuf.Empty) returns (PongReply) {
    option(google.api.http) = {
      get:  "/pong"
    };
  };
}
```

## Examples

### Ping Pong Server

This is a simple server that tests the functionality of the gateway. It demonstrates how a service is put together.

```
# start the pingpong server
make example/pingpong/run-server

# run the grpc client
make example/pingpong/run-grpc-client
INFO[0000] message:"Hello John"

# run the rest client
make example/pingpong/run-rest-client
{"message":"Hello John"}%

# test the custom rest routes
http --verify=no https://localhost:5000/custom/rest    
HTTP/1.1 200 OK
Content-Length: 8
Content-Type: text/plain; charset=utf-8
Date: Fri, 24 Nov 2017 10:28:27 GMT

pingpong

```

## Kudos

The implementation is inspired by [Brandon Philips GRPC example](https://github.com/philips/grpc-gateway-example/blob/master/cmd/serve.go). He explained the concepts in [Take a REST with HTTP/2, Protobufs, and Swagger](https://coreos.com/blog/grpc-protobufs-swagger.html)
