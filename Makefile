GO   := GO15VENDOREXPERIMENT=1 go
pkgs  = $(shell $(GO) list ./... | grep -v /vendor/)

all: proto format build

get:
	# govendor
	go get -u -v github.com/kardianos/govendor
	# proto
	go get -u -v github.com/golang/protobuf/protoc-gen-go
	# gogoproto
	go get -u -v github.com/gogo/protobuf/proto
	go get -u -v github.com/gogo/protobuf/protoc-gen-gogo
	go get -u -v github.com/gogo/protobuf/gogoproto
	go get -u -v github.com/gogo/protobuf/protoc-gen-gogoslick
	go get -u -v github.com/gogo/protobuf/protoc-gen-gogofast
	# grpc-gateway
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	# docs
	# npm install -g bootprint bootprint-openapi html-inline

docs:
	@mkdir -p ./protocol/docs
	@echo ">> generating swagger from protobuf"
	@protoc -I /usr/local/include -I ./protocol/ \
		-I ${GOPATH}/src \
		-I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--swagger_out=logtostderr=true:./protocol/docs \
		./protocol/password.proto
	#bootprint openapi ./protocol/docs/*.swagger.json ./protocol/docs/

proto:
	@mkdir -p ./protocol/password
	@echo ">> protobuf"
	@echo " >   bindings"
	@protoc -I ./protocol/ -I /usr/local/include \
	 			 -I ${GOPATH}/src \
				 -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
				 --gogofast_out=plugins=grpc:./protocol/password \
				 ./protocol/*.proto
	@echo " >   gateway"
	@protoc -I ./protocol/ -I /usr/local/include \
		 		 -I ${GOPATH}/src \
				 -I ${GOPATH}/src/github.com/gogo/protobuf \
				 -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
				 --grpc-gateway_out=logtostderr=true:./protocol/password \
				 ./protocol/*.proto

test:
	@echo ">> running tests"
	@$(GO) test -short $(pkgs)

vendor:
	govendor init
	govendor add -v +external
	govendor update -v +vendor

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

build:
	@echo ">> building binaries"
	@./ci/build.sh

pack: build
	@echo ">> packing all binaries"
	@upx -9 bin/*

docker: pack
	@docker build -t password:$(shell cat version/VERSION)-$(shell git rev-parse --short HEAD) .

docker-release: docker
	@docker tag password:$(shell cat version/VERSION)-$(shell git rev-parse --short HEAD) password:$(shell cat version/VERSION)
	@docker tag password:$(shell cat version/VERSION)-$(shell git rev-parse --short HEAD) password:latest

.PHONY: all get proto format build test vet docker assets vendor
