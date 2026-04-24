.PHONY: build clean help

BINARY := youtube-randomizer

build:
	go build -o $(BINARY) .

help:
	go run . -h

clean:
	rm -f $(BINARY)
