.PHONY: build build-tusk build-tuskd lint test clean fmt

build: build-tusk build-tuskd

build-tusk:
	go build -o tusk ./cmd/tusk

build-tuskd:
	go build -o tuskd ./cmd/tuskd

lint:
	golangci-lint run ./... || true

test:
ifeq ($(TUSK_NESTED),1)
	@echo "nested test run - skipping"
else
	TUSK_NESTED=1 go test ./...
endif

clean:
	rm -f tusk tuskd

fmt:
	go fmt ./...
