ifndef GOBIN
  GOBIN=go
endif

all: deps
		$(GOBIN) install

deps:
		$(GOBIN) get github.com/aws/aws-sdk-go
test:
		$(GOBIN) test -v ./...
