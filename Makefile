SOURCES := $(shell find . -name '*.go')

pggateway: $(SOURCES)
	go build -o pggateway cmd/pggateway/main.go

clean:
	rm -f ./pggateway
.PHONY: clean


run:
	go run cmd/pggateway/main.go
.PHONY: run
