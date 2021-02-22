PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := ${HOME}/bin/protobuf/bin/protoc

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

resources/dcs_model.pb.go: proto/dcs_model.proto | $(PROTOC_GEN_GO)
	 $(PROTOC) -I proto --go_out=resources proto/dcs_model.proto

protos: resources/dcs_model.pb.go

.DEFAULT_GOAL := protos
.PHONY:	protos
