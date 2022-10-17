GOOS := darwin
GOARCH := amd64

APP_NAME := go-crawer
RELEASE := dev

BUILD_TIME = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

BIN = ./bin
SPBS_ROOT_PATH = ../..

all: clean build

build: build-admin build-go-crawer

build-admin:
	@echo "===>> building admin"
	@CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} CC=${CGO_CC} CXX=${CGO_CXX} \
	  go build -ldflags='${GO_LD_FLAGS}' \
	  -o ${BIN}/admin ./admin/admin.go
	@echo "===>> build admin success"
	cp ./admin/admin_config.yaml ${BIN}

build-go-crawer:
	@echo "===>> building go-crawer"
	@CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} CC=${CGO_CC} CXX=${CGO_CXX} \
	  go build -ldflags='${GO_LD_FLAGS}' \
	  -o ${BIN}/go-crawer ./server.go
	@echo "===>> build go-crawer success"
	cp ./config/config.yaml ${BIN}
	cp -r ./parser/rules/ ${BIN}/rules

clean:
	rm -fr dist

lint:
	golint ./...

mod-tidy:
	go mod tidy

