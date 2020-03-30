export

.DEFAULT_GOAL := all


## Install dependencies and build Docker image
all: deps protobuf build


## Install dependencies
deps:
	$(info Installing dependencies)
	@go mod tidy


## Build the project binary
build:
	$(info Building project)
	@CGO_ENABLED=0 GOOS=linux go build -a \
		-installsuffix cgo -o xyz \
		./cmd/main.go


## Run the service
run:
	$(info Running the service)
	@go run cmd/main.go


## generate the gRPC code 
protobuf:
	$(info Creating Go protobuf)
	@protoc -I=pb -I=${GOPATH}/src \
		-I=${GOPATH}/src/github.com/google/protobuf/src/ \
		-I=${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway \
		-I=${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
        --swagger_out=logtostderr=true:docs \
		--grpc-gateway_out=logtostderr=true:pb \
		--go_out=plugins=grpc,paths=source_relative:pb \
		pb/*.proto


## Run the API tests
test:
	$(info Running tests)
	@./test.sh


.PHONY: all build deps protobuf run test
