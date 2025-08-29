APP     := imsi-bot
ARCH    ?= arm64
OS      ?= linux
ENTRY   := ./cmd/main.go

LDVERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS   := -s -w -buildid= -X 'main.Version=$(LDVERSION)'

OUT   := release/$(APP)-$(OS)-$(ARCH)

# --- SSH / VPS ---
SSH_HOST ?= 7936d3aa723.vps.myjino.ru
SSH_PORT ?= 49236
SSH_USER ?= root
SSH      := $(SSH_USER)@$(SSH_HOST)
SSH_CMD  = ssh -p $(SSH_PORT) $(SSH)
SCP_CMD  = scp -P $(SSH_PORT)

# --- PG туннель ---
TUNNEL_LPORT ?= 5436     # локальный порт на Маке
PG_RPORT     ?= 5436     # порт Postgres на VPS
SSH_SOCK     := /tmp/$(APP)-ssh.sock

# --- ENV для локального запуска через туннель ---
POSTGRES_HOST ?= 127.0.0.1
POSTGRES_PORT ?= $(TUNNEL_LPORT)

.PHONY: tidy build clean \
        build-linux-amd64 build-linux-arm64 \
        build-windows-amd64 build-windows-arm64 build-darwin-arm64 \
        tunnel tunnel-stop tunnel-status run

tidy:
	go mod tidy

build:
	@mkdir -p release
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 \
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(OUT) $(ENTRY)
	@echo "Built: $(OUT)"

build-linux-amd64:  ; $(MAKE) build OS=linux  ARCH=amd64
build-linux-arm64:  ; $(MAKE) build OS=linux  ARCH=arm64
build-windows-amd64:; $(MAKE) build OS=windows ARCH=amd64
build-windows-arm64:; $(MAKE) build OS=windows ARCH=arm64
build-darwin-arm64: ; $(MAKE) build OS=darwin ARCH=arm64

clean:
	rm -rf release
