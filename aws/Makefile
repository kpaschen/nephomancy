PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := ${HOME}/bin/protobuf/bin/protoc

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

resources/awsec2_model.pb.go: proto/awsec2_model.proto | $(PROTOC_GEN_GO)
	 $(PROTOC) -I proto --go_out=resources proto/awsec2_model.proto

protos: resources/awsec2_model.pb.go

.DEFAULT_GOAL := protos
.PHONY:	protos
