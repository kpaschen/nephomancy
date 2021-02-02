PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC := ${HOME}/bin/protobuf/bin/protoc

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go

assets/gcloud_model.pb.go: proto/gcloud_model.proto | $(PROTOC_GEN_GO)
	 $(PROTOC) -I proto --go_out=assets proto/gcloud_model.proto

protos: assets/gcloud_model.pb.go

.DEFAULT_GOAL := protos
.PHONY:	protos
