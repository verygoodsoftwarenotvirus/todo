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
	docker run --env GO111MODULE=on --volume `pwd`:`pwd` --workdir=`pwd` --workdir=`pwd` golang:latest /bin/sh -c "pwd; ls -Al; go mod vendor"

.PHONY: revendor
revendor:
	rm -rf vendor go.{mod,sum}
	GO111MODULE=on go mod init
	$(MAKE) vendor

.PHONY: prerequisites
prerequisites:
	$(MAKE) vendor $(SERVER_PRIV_KEY) $(SERVER_CERT_KEY) $(CLIENT_PRIV_KEY) $(CLIENT_CERT_KEY)

dev_files/certs/client/key.pem dev_files/certs/client/cert.pem:
	mkdir -p dev_files/certs/client
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com" -keyout dev_files/certs/client/key.pem -out dev_files/certs/client/cert.pem

dev_files/certs/server/key.pem dev_files/certs/server/cert.pem:
	mkdir -p dev_files/certs/server
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com" -keyout dev_files/certs/server/key.pem -out dev_files/certs/server/cert.pem

## Test things
example.db:
	go run tools/db-bootstrap/main.go

$(COVERAGE_OUT):
	./scripts/coverage.sh

.PHONY: ci-coverage
ci-coverage:
	docker build --tag coverage-todo:latest --file dockerfiles/coverage.Dockerfile .
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` coverage-todo:latest

.PHONY: integration-tests
integration-tests:
	docker-compose --file compose-files/integration-tests.yaml up --build --remove-orphans --force-recreate --abort-on-container-exit

.PHONY: test
test:
	for pkg in $(TESTABLE_PACKAGES); do \
		go test -cover -v -count 5 $$pkg; \
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
