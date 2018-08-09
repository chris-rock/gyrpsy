# API Gateway Benchmark

## Test Setup

<a target="_blank" href="docs/setup.png"><img src="docs/setup.png" alt="test setup" title="Test Setup" style="width:60%;"></a>

For all setups, we are going to use a very simple GRPC service definition, since the focus is to understand the impact of the service assembly.

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

The service implementation for all setups are the same and located in `/api`. All REST endpoint are auto-generated via the [GRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway).

## Test Results

```
go test -run=^$ github.com/chris-rock/gyrpsy/bench -bench=. -benchtime 5s
goos: darwin
goarch: amd64
pkg: github.com/chris-rock/gyrpsy/bench
BenchmarkGoGatewayGRPC/pingpong_proto_ingestion-2         	   30000	    301696 ns/op
BenchmarkDirectGrpc/pingpong_proto_ingestion-2            	   50000	    151657 ns/op
BenchmarkNginxTLSGrpc/pingpong_proto_ingestion-2          	   10000	   1972509 ns/op
BenchmarkNginxGrpc/pingpong_proto_ingestion-2             	   10000	    896205 ns/op
BenchmarkGoGatewayRest/pingpong_json_ingestion_-2         	    2000	  17835365 ns/op
BenchmarkDirectRest/pingpong_json_ingestion_-2            	    2000	  18514855 ns/op
BenchmarkDirectRestHTTP/pingpong_json_ingestion_-2        	    5000	   7968036 ns/op
BenchmarkNginxRestHTTPS/pingpong_json_ingestion_-2        	    2000	   7949977 ns/op

PASS
ok  	github.com/chris-rock/gyrpsy/bench	179.920s
```

This benchmark also demonstrates the obvious: Try to avoid additional network hops for your incoming requests if possible.

## Kudos

The implementation of the Go Gateway was inspired by [Brandon Philips GRPC example](https://github.com/philips/grpc-gateway-example/blob/master/cmd/serve.go). He explained the concepts in [Take a REST with HTTP/2, Protobufs, and Swagger](https://coreos.com/blog/grpc-protobufs-swagger.html)


## References

- [Introducing gRPC Support with NGINX 1.13.10](https://www.nginx.com/blog/nginx-1-13-10-grpc/)
- [Deploying NGINX Plus as an API Gateway, Part 1](https://www.nginx.com/blog/deploying-nginx-plus-as-an-api-gateway-part-1/)
- [Deploying NGINX Plus as an API Gateway, Part 2: Protecting Backend Services](https://www.nginx.com/blog/deploying-nginx-plus-as-an-api-gateway-part-2-protecting-backend-services/)