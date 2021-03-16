GOLINT := $(GOPATH)/bin/golint

$(GOLINT):
	go get -u golang.org/x/lint/golint

common:
	$(MAKE) -C common/

gcloud:
	$(MAKE) -C gcloud/

dcs:
	$(MAKE) -C dcs/

aws:
	$(MAKE) -C aws/

nephomancy: common gcloud dcs
	go build

test: nephomancy
	go test -v -cover nephomancy/...

lint: nephomancy | $(GOLINT)
	go fmt -n ./...
	$(GOLINT) ./...
	go vet ./...

.PHONY: common gcloud dcs aws

.DEFAULT_GOAL := nephomancy

