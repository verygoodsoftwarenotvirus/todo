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

.PHONY: dockercide
dockercide:
	docker system prune --force --all --volumes

## Project prerequisites
vendor:
	docker run --env GO111MODULE=on --volume `pwd`:`pwd` --workdir=`pwd` golang:latest /bin/sh -c "go mod vendor"

.PHONY: revendor
revendor:
	rm -rf vendor go.{mod,sum}
	GO111MODULE=on go mod init
	$(MAKE) vendor

.PHONY: dev-tools
dev-tools:
	go get -u github.com/google/wire/cmd/wire

.PHONY: wire
wire:
	wire gen gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

.PHONY: prerequisites
prerequisites: dev-tools
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
integration-tests: wire
	docker-compose --file compose-files/integration-tests.yaml up --always-recreate-deps --build --remove-orphans --force-recreate --abort-on-container-exit

.PHONY: debug-integration-tests
debug-integration-tests: wire # literally the same except it won't exit
	docker-compose --file compose-files/integration-tests.yaml up --always-recreate-deps --build --remove-orphans --force-recreate

## Docker things

.PHONY: docker-image
docker-image: prerequisites wire-gen-server
	docker build --tag todo:latest --file dockerfiles/server.Dockerfile .

## Running

.PHONY: run
run: docker-image
	docker-compose --file compose-files/docker-compose.yaml up --always-recreate-deps --build --remove-orphans --abort-on-container-exit --force-recreate

.PHONY: run-local
run-local:
	go run cmd/server/v1/main.go

.PHONY: run-local-integration-server
run-local-integration-server:
	docker-compose --file compose-files/integration-debug.yaml up --always-recreate-deps --build --remove-orphans --force-recreate
