
example/pingpong/run-server:
	@go run examples/pingpong/server/main.go

example/pingpong/run-grpc-client:
	@go run examples/pingpong/client/grpc/main.go

example/pingpong/run-rest-client:
	@go run examples/pingpong/client/rest/main.go

example/pingpong/proto:
	@./examples/pingpong/proto.sh

unit:
	@go test -v $(shell go list ./... | grep -v '/vendor/') -cover
