.PHONY: build help install uninstall clean

BINARY := youtube-randomizer
PREFIX ?= /usr/local
BIN    := $(PREFIX)/bin/$(BINARY)
UNIT   := /etc/systemd/system/$(BINARY).service
ENVFILE := /etc/$(BINARY).env

build:
	go build -o $(BINARY) .

help:
	go run . -h

# Install a prebuilt binary + systemd unit + placeholder env file. Idempotent.
# Expects `make build` to have been run first (so we don't need Go under sudo).
install:
	@test -x ./$(BINARY) || { echo "error: ./$(BINARY) not found. Run 'make build' first (without sudo)."; exit 1; }
	install -m 755 ./$(BINARY) $(BIN)
	id -u $(BINARY) >/dev/null 2>&1 || useradd --system --no-create-home --shell /usr/sbin/nologin $(BINARY)
	test -f $(ENVFILE) || install -m 600 -o root -g root .env.example $(ENVFILE)
	install -m 644 systemd/$(BINARY).service $(UNIT)
	systemctl daemon-reload
	@echo
	@echo "Installed. Edit $(ENVFILE) (set YT_API_KEY), then:"
	@echo "  sudo systemctl enable --now $(BINARY)"

uninstall:
	-systemctl disable --now $(BINARY)
	-rm -f $(BIN) $(UNIT)
	systemctl daemon-reload
	@echo "Left $(ENVFILE) and user '$(BINARY)' in place. Remove manually if you want."

clean:
	rm -f $(BINARY)
