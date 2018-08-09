.PHONY: bench

start/go-gateway:
	@go run gateway/go/main.go

start/go-nginx:
	# run nginx in foreground, ensure config is placed properly eg. /usr/local/etc/nginx/nginx.conf (macos)
	@nginx -g 'daemon off;'

start/service-grpc:
	@go run services/pp_grpc/main.go

start/service-rest:
	@go run services/pp_rest/main.go

client/grpc:
	@go run bench/client/grpc/main.go

client/rest:
	@go run bench/client/rest/main.go

bench:
	go test -v -benchmem -run=github.com/chris-rock/gyrpsy/bench -bench .
	
unit:
	@go test -v $(shell go list ./... | grep -v '/vendor/') -cover

