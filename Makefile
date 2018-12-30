GOPATH            := $(GOPATH)
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
	docker run --env GO111MODULE=on --volume `pwd`:`pwd` --workdir=`pwd` --workdir=`pwd` golang:latest /bin/sh -c "go mod vendor"

.PHONY: revendor
revendor:
	rm -rf vendor go.{mod,sum}
	GO111MODULE=on go mod init
	$(MAKE) vendor

.PHONY: dev-tools
dev-tools:
	go get -u github.com/jsha/minica

.PHONY: prerequisites
prerequisites:
	$(MAKE) vendor $(SERVER_PRIV_KEY) $(SERVER_CERT_KEY) $(CLIENT_PRIV_KEY) $(CLIENT_CERT_KEY)

dev_files/certs/client/key.pem dev_files/certs/client/cert.pem:
	mkdir -p dev_files/certs/client
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=localhost" -keyout dev_files/certs/client/key.pem -out dev_files/certs/client/cert.pem

dev_files/certs/server/key.pem dev_files/certs/server/cert.pem:
	mkdir -p dev_files/certs/server
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=localhost" -keyout dev_files/certs/server/key.pem -out dev_files/certs/server/cert.pem

## Test things
example.db:
	go run tests/integration/v1/db_bootstrap/main.go

$(COVERAGE_OUT):
	./scripts/coverage.sh

.PHONY: test
test:
	docker build --tag coverage-todo:latest --file dockerfiles/coverage.Dockerfile .
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` coverage-todo:latest

.PHONY: integration-tests
integration-tests:
	docker-compose --file compose-files/integration-tests.yaml up --build --remove-orphans --abort-on-container-exit --force-recreate

## Docker things

.PHONY: docker-image
docker-image: prerequisites
	docker build --tag todo:latest --file dockerfiles/server.Dockerfile .

## Running

.PHONY: run
run: docker-image
	docker run --rm --publish 443:443 todo:latest

.PHONY: run-local
run-local:
	go run cmd/server/v1/main.go

.PHONY: run-local-integration-server
run-local-integration-server:
	docker build --tag dev-todo:latest --file dockerfiles/integration-server.Dockerfile .
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` --publish=443 dev-todo:latest
