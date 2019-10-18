GOOS ?= linux
GOARCH ?= amd64
SRV = $(notdir $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
SRV_WORKER = $(notdir $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))-worker
PROJECT = github.com/makerdao/${SRV}
TAG ?= latest
BUILD ?= `git rev-parse --short HEAD`
PORT ?= 5001
CA_DIR ?= certs
PWD ?= $(pwd)
REGISTRY ?= makerdao/

build: vendor lint certs
	@echo "+ $@ ${GOOS}"
	@CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -a -installsuffix cgo \
		-o bin/${GOOS}-${GOARCH}/service ${PROJECT}/cmd/rpc
	@CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -a -installsuffix cgo \
        -o bin/${GOOS}-${GOARCH}/worker ${PROJECT}/cmd/worker
.PHONY: build

vendor:
	@echo "+ $@"
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor
.PHONY: vendor

run: build
	@echo "+ $@ ${GOOS}"
	@bin/${GOOS}-${GOARCH}/service
.PHONY: run

run-worker: build
	@echo "+ $@ ${GOOS}"
	@bin/${GOOS}-${GOARCH}/worker
.PHONY: run-worker

test:
	@echo "+ $@"
	@mkdir ${PWD}/.testdir && echo "testdir created" || echo "testdir already exists"
	@go test -count=1 -parallel 1 ./...
.PHONY: test

lint:
	@echo "+ $@"
	@docker run --rm -i  \
		-v ${GOPATH}/src/${PROJECT}:/go/src/${PROJECT} \
		-w /go/src/${PROJECT} golangci/golangci-lint:v1.12 golangci-lint run --enable-all --skip-dirs vendor,version,pkg/gen ./...
.PHONY: lint

docker-push:
	@echo "Pushing docker images"
	@docker push ${REGISTRY}${SRV}-base:latest
	@docker push ${REGISTRY}${SRV}:${TAG}
	@docker push ${REGISTRY}${SRV_WORKER}:${TAG}
.PHONY: docker-push

build-image: build
	@echo "+ $@"
	@docker build \
		-t ${REGISTRY}${SRV}:${BUILD} \
		-t ${REGISTRY}${SRV}:${TAG} .
.PHONY: build-image

build-worker-image: build
	@echo "+ $@"
	@docker build \
		-t ${REGISTRY}${SRV_WORKER}:${BUILD} \
		-t ${REGISTRY}${SRV_WORKER}:${TAG} \
		-f Dockerfile.worker .
.PHONY: build-worker-image

build-base-image:
	@echo "+ $@"
	@docker build \
		-t ${REGISTRY}${SRV}-base:${TAG} \
		-f base.Dockerfile .
.PHONY: build-image

stop-image:
	@echo "+ $@"
	@docker stop ${SRV} && echo "container stoped" || echo "container is not runned"
	@docker rm -f ${SRV} && echo "container removed" || echo "container not exists"
.PHONY: build-image

run-image: stop-image build-image
	@echo "+ $@"
	@docker run -d -p ${PORT}:${PORT} \
		-e TCD_PORT='${PORT}' \
		-v ~/.ssh:/root/.ssh \
		--name=${SRV} ${REGISTRY}${SRV}:${TAG}
.PHONY: run-image

run-image-local: stop-image build-image
	@echo "+ $@"
	@docker run -d -p ${PORT}:${PORT} \
	    -e TCD_GATEWAY='host=host.docker.internal' \
		-e TCD_PORT='${PORT}' \
		--name=${SRV} ${REGISTRY}${SRV}:${TAG}
.PHONY: run-image-local


logs:
	@echo "+ $@"
	@docker logs -f ${SRV}
.PHONY: image-logs

certs:
ifeq ("$(wildcard $(CA_DIR)/ca-certificates.crt)","")
	@echo "+ $@"
	@docker run --name ${SRV}-certs -d alpine:latest sh -c "apk --update upgrade && apk add ca-certificates && update-ca-certificates"
	@docker wait ${SRV}-certs
	@mkdir -p ${CA_DIR}
	@docker cp ${SRV}-certs:/etc/ssl/certs/ca-certificates.crt ${CA_DIR}
	@docker rm -f ${SRV}-certs
endif
.PHONY: certs
