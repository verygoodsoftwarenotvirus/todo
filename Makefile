GOPATH            := $(GOPATH)
GIT_HASH          := $(shell git describe --tags --always --dirty)
BUILD_TIME        := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
TESTABLE_PACKAGES := $(shell go list gitlab.com/verygoodsoftwarenotvirus/todo/...)

SERVER_PRIV_KEY := dev_files/certs/server/key.pem
SERVER_CERT_KEY := dev_files/certs/server/cert.pem
CLIENT_PRIV_KEY := dev_files/certs/client/key.pem
CLIENT_CERT_KEY := dev_files/certs/client/cert.pem

## generic make stuff

.PHONY: clean
clean:
	rm example.db

## Project prerequisites

vendor:
	GO111MODULE=on go mod init
	GO111MODULE=on go mod vendor

.PHONY: revendor
revendor:
	rm -rf vendor go.{mod,sum}
	$(MAKE) vendor

.PHONY: prerequisites
prerequisites: vendor $(SERVER_PRIV_KEY) $(SERVER_CERT_KEY) $(CLIENT_PRIV_KEY) $(CLIENT_CERT_KEY)

dev_files/certs/client/key.pem dev_files/certs/client/cert.pem:
	mkdir -p dev_files/certs/client
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -keyout dev_files/certs/client/key.pem -out dev_files/certs/client/cert.pem

dev_files/certs/server/key.pem dev_files/certs/server/cert.pem:
	mkdir -p dev_files/certs/server
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -keyout dev_files/certs/server/key.pem -out dev_files/certs/server/cert.pem

## Test things

.PHONY: coverage
coverage:
	if [ -f coverage.out ]; then rm coverage.out; fi
	echo "mode: set" > coverage.out

	for pkg in $(TESTABLE_PACKAGES); do \
		set -e; go test -coverprofile=profile.out -v -race $$pkg; \
		cat profile.out | grep -v "mode: atomic" >> coverage.out; \
	done
	rm profile.out

.PHONY: ci-coverage
ci-coverage:
	go test $(TESTABLE_PACKAGES) -v -coverprofile=profile.out

example.db:
	go run tools/db-bootstrap/main.go

.PHONY: integration-tests
integration-tests:
	docker-compose --file compose-files/integration-tests.yaml up --build --remove-orphans --force-recreate --abort-on-container-exit

## Docker things

.PHONY: docker-image
docker-image: prerequisites
	docker build --tag todo:latest --file dockerfiles/server.Dockerfile .

## Running

.PHONY: run
run: docker-image
	docker run --rm --publish 1443:443 todo:latest

.PHONY: run-local
run-local:
	go run cmd/server/main.go
