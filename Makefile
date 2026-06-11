.PHONY: build run clean install fmt test release-snapshot lint

build:
	go build -o permission-extractor-tui ./cmd

run: build
	./permission-extractor-tui

clean:
	rm -f permission-extractor-tui
	rm -f *.csv
	rm -rf dist/

install: build
	cp permission-extractor-tui /usr/local/bin/

fmt:
	go fmt ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

# Release snapshot (local test, no publish)
release-snapshot:
	goreleaser release --snapshot --clean

# Full release (requires tag)
release:
	goreleaser release --clean
