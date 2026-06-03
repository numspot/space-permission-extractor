.PHONY: build run clean install

build:
	go build -o permission-extractor-tui ./cmd

run: build
	./permission-extractor-tui

clean:
	rm -f permission-extractor-tui
	rm -f *.csv

install: build
	cp permission-extractor-tui /usr/local/bin/

fmt:
	go fmt ./...

test:
	go test ./...
