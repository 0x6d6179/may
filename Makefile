VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/0x6d6179/may/internal/version.Version=$(VERSION)
BINARY  := may
CMD     := ./cmd/may

.PHONY: build test vet lint install clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

test:
	go test ./... -race -count=1

vet:
	go vet ./...

lint:
	golangci-lint run

install:
	go install -ldflags "$(LDFLAGS)" $(CMD)

clean:
	rm -f $(BINARY)
