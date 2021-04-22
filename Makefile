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

nephomancy: common gcloud dcs aws
	go build

test: nephomancy
	go test -v -cover nephomancy/...

integration_tests: nephomancy
	go test -v -cover --tags=integration nephomancy/...

lint: nephomancy | $(GOLINT)
	go fmt -n ./...
	$(GOLINT) ./...
	go vet ./...

.PHONY: common gcloud dcs aws

.DEFAULT_GOAL := nephomancy

