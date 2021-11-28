genpb: 
	protoc --go_out=serverpb --go_opt=paths=source_relative \
    --go-grpc_out=serverpb --go-grpc_opt=paths=source_relative \
    proto/server.proto
gengrpc:
	protoc -I . --grpc-gateway_out ./serverpb \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=source_relative \
    --grpc-gateway_opt generate_unbound_methods=true \
    proto/server.proto
genopenapi:
	protoc -I . --openapiv2_out ./swaggerui \
    --openapiv2_opt logtostderr=true \
    proto/server.proto
generate:
	buf generate
BUF_VERSION:=0.43.2

run: 
	go run main.go
	
certt:
	cd cert; ./gen.sh; cd ..

.PHONY: gen clean server

install:
	go install \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	curl -sSL \
    	"https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(shell uname -s)-$(shell uname -m)" \
    	-o "$(shell go env GOPATH)/bin/buf" && \
  	chmod +x "$(shell go env GOPATH)/bin/buf"