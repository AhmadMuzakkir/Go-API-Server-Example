PROJECT?=github.com/ahmadmuzakkir/go-sample-api-server-structure
APP?=server
RELEASE?=0.0.1
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

GOOS?=darwin
GOARCH?=amd64

.PHONY: clean
clean:
	rm -f ${APP}

.PHONY: build
build: clean
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
		-ldflags "-X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o ${APP} ./cmd/server/main.go

.PHONY: run
run: build
	./server

.PHONY: run-docker
run-docker:
	docker-compose down
	./scripts/run.docker-compose.sh

.PHONY: clean-docker
clean-docker:
	docker-compose down --volumes