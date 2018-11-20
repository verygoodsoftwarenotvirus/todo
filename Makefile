GOPATH            := $(GOPATH)
GIT_HASH          := $(shell git describe --tags --always --dirty)
BUILD_TIME        := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
TESTABLE_PACKAGES := $(shell go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|tools)')
COVERAGE_OUT      := coverage.out

SERVER_PRIV_KEY := dev_files/certs/server/key.pem
SERVER_CERT_KEY := dev_files/certs/server/cert.pem
CLIENT_PRIV_KEY := dev_files/certs/client/key.pem
CLIENT_CERT_KEY := dev_files/certs/client/cert.pem

## generic make stuff

.PHONY: clean
clean:
	rm -f $(COVERAGE_OUT)
	rm -f example.db

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

$(COVERAGE_OUT):
	echo "mode: set" > $(COVERAGE_OUT)

	for pkg in $(TESTABLE_PACKAGES); do \
		go test -coverprofile=profile.out -v -race -count 5 $$pkg; \
		cat profile.out | grep -v "mode: atomic" >> $(COVERAGE_OUT); \
	done
	rm profile.out

example.db:
	go run tools/db-bootstrap/main.go

.PHONY: integration-tests
integration-tests:
	docker-compose --file compose-files/integration-tests.yaml up --build --remove-orphans --force-recreate --abort-on-container-exit

.PHONY: test
test:
	for pkg in $(TESTABLE_PACKAGES); do \
		go test -cover -v -race -count 5 $$pkg; \
	done

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
