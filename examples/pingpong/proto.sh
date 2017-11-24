#!/bin/bash

# service
protoc -I. \
  -I$GOPATH/src \
  -I$PWD/vendor \
  -I$PWD/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --go_out=plugins=grpc:. \
  examples/pingpong/proto/*.proto

# generates grpc gateway
protoc -I. \
  -I$GOPATH/src \
  -I$PWD/vendor \
  -I$PWD/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --grpc-gateway_out=request_context=true,logtostderr=true:. \
  examples/pingpong/proto/*.proto
