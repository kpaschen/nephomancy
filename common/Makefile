# Makefile based on blog.gopheracademy.com/advent-2017/make

PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := ${HOME}/bin/protobuf/bin/protoc

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

resources/model.pb.go: proto/model.proto | $(PROTOC_GEN_GO)
	 $(PROTOC) -I proto --go_out=resources proto/model.proto

protos: resources/model.pb.go

.DEFAULT_GOAL := protos
.PHONY:	protos
