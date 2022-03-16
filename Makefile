APP?=edgeproxy
RELEASE?=0.0.0
DOCKER_IMAGE_REPO=segator
GOOS?=linux
GOARCH?=amd64
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
SERVER_PORT?=9180
clean:
	rm -f ${APP}
#CGO_ENABLED=0
build: clean
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
		-ldflags "-X 'edgeproxy/version.release=${RELEASE}' \
		-X 'edgeproxy/version.commit=${COMMIT}' -X 'edgeproxy/version.buildTime=${BUILD_TIME}'" \
		-o ${APP}
	chmod +x ${APP}

run: build
	PORT=${PORT} ./${APP}

test:
	go test -v -race ./...

container: build
	docker build -t $(DOCKER_IMAGE_REPO)/$(APP):$(RELEASE) .

run_server: container
	docker stop $(DOCKER_IMAGE_REPO)/$(APP):$(RELEASE) || true && docker rm $(DOCKER_IMAGE_REPO)/$(APP):$(RELEASE) || true
	docker run --name ${APP} -p ${SERVER_PORT}:${SERVER_PORT} --rm $(DOCKER_IMAGE_REPO)/$(APP):$(RELEASE)
