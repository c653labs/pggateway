SOURCES := $(shell find . -name '*.go')

pggateway: $(SOURCES)
	CGO_ENABLED=0 go build -o pggateway -a -ldflags "-s -w" cmd/pggateway/main.go

clean:
	rm -f ./pggateway
.PHONY: clean

run:
	go run cmd/pggateway/main.go
.PHONY: run

debug:
	dlv debug cmd/pggateway/main.go
.PHONY: debug

test:
	@go test -v  ./...
