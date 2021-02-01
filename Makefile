GOLINT := $(GOPATH)/bin/golint

$(GOLINT):
	go get -u golang.org/x/lint/golint

common:
	$(MAKE) -C common/

gcloud:
	$(MAKE) -C gcloud/

nephomancy: common gcloud
	go build

test: nephomancy
	go test -v -cover nephomancy/...

lint: nephomancy | $(GOLINT)
	go fmt -n ./...
	$(GOLINT) ./...
	go vet ./...

.PHONY: common

.PHONY: nephomancy

.DEFAULT_GOAL := nephomancy

