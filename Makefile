PWD           := $(shell pwd)
GOPATH        := $(GOPATH)
ARTIFACTS_DIR := artifacts
COVERAGE_OUT  := $(ARTIFACTS_DIR)/coverage.out
CONFIG_DIR    := config_files
GO_FORMAT     := gofmt -s -w

SERVER_DOCKER_IMAGE_NAME := todo-server
SERVER_DOCKER_REPO_NAME  := docker.io/verygoodsoftwarenotvirus/$(SERVER_DOCKER_IMAGE_NAME)

$(ARTIFACTS_DIR):
	mkdir -p $(ARTIFACTS_DIR)

## dependency injection

.PHONY: wire-clean
wire-clean:
	rm -f cmd/server/v1/wire_gen.go

.PHONY: wire
wire:
	wire gen gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

.PHONY: rewire
rewire: wire-clean wire

## Go-specific prerequisite stuff

.PHONY: dev-tools
dev-tools:
	GO111MODULE=off go get -u github.com/google/wire/cmd/wire
	GO111MODULE=off go get -u github.com/axw/gocov/gocov

.PHONY: vendor-clean
vendor-clean:
	rm -rf vendor go.sum

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor

.PHONY: revendor
revendor: vendor-clean vendor

## Config

clean-configs:
	rm -rf $(CONFIG_DIR)

$(CONFIG_DIR):
	mkdir -p $(CONFIG_DIR)
	go run cmd/config_gen/v1/main.go

## Testing things

.PHONY: lint
lint:
	@docker pull golangci/golangci-lint:latest
	docker run \
		--rm \
		--volume `pwd`:`pwd` \
		--workdir=`pwd` \
		--env=GO111MODULE=on \
		golangci/golangci-lint:latest golangci-lint run --config=.golangci.yml ./...

$(COVERAGE_OUT): $(ARTIFACTS_DIR)
	set -ex; \
	echo "mode: set" > $(COVERAGE_OUT);
	for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|mock)'`; do \
		go test -coverprofile=profile.out -v -count 5 -race -failfast $$pkg; \
		if [ $$? -ne 0 ]; then exit 1; fi; \
		cat profile.out | grep -v "mode: atomic" >> $(COVERAGE_OUT); \
	rm -f profile.out; \
	done || exit 1
	gocov convert $(COVERAGE_OUT) | gocov report

.PHONY: quicktest # basically the same as coverage.out, only running once instead of with `-count` set
quicktest: $(ARTIFACTS_DIR)
	@set -ex; \
	echo "mode: set" > $(COVERAGE_OUT);
	for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|mock)'`; do \
		go test -coverprofile=profile.out -race -failfast $$pkg; \
		if [ $$? -ne 0 ]; then exit 1; fi; \
		cat profile.out | grep -v "mode: atomic" >> $(COVERAGE_OUT); \
	rm -f profile.out; \
	done || exit 1
	gocov convert $(COVERAGE_OUT) | gocov report

.PHONY: coverage-clean
coverage-clean:
	@rm -f $(COVERAGE_OUT) profile.out;

.PHONY: coverage
coverage: coverage-clean $(COVERAGE_OUT)

.PHONY: test
test:
	docker build --tag coverage-$(SERVER_DOCKER_IMAGE_NAME):latest --file dockerfiles/coverage.Dockerfile .
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` coverage-$(SERVER_DOCKER_IMAGE_NAME):latest

.PHONY: format
format:
	for file in `find $(PWD) -name '*.go'`; do $(GO_FORMAT) $$file; done

.PHONY: frontend-tests
frontend-tests:
	docker-compose --file compose-files/frontend-tests.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

## Integration tests

.PHONY: lintegration-tests # this is just a handy lil' helper I use sometimes
lintegration-tests: integration-tests lint

.PHONY: integration-tests
integration-tests: integration-tests-postgres integration-tests-sqlite integration-tests-mariadb

.PHONY: integration-tests-postgres
integration-tests-postgres:
	docker-compose --file compose-files/integration-tests-postgres.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: integration-tests-sqlite
integration-tests-sqlite:
	docker-compose --file compose-files/integration-tests-sqlite.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: integration-tests-mariadb
integration-tests-mariadb:
	docker-compose --file compose-files/integration-tests-mariadb.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: integration-coverage
integration-coverage:
	@# big thanks to https://blog.cloudflare.com/go-coverage-with-external-tests/
	rm -f ./artifacts/integration-coverage.out
	mkdir -p ./artifacts
	docker-compose --file compose-files/integration-coverage.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit
	go tool cover -html=./artifacts/integration-coverage.out

## Load tests

.PHONY: load-tests
load-tests: load-tests-postgres load-tests-sqlite load-tests-mariadb

.PHONY: load-tests-postgres
load-tests-postgres:
	docker-compose --file compose-files/load-tests-postgres.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: load-tests-sqlite
load-tests-sqlite:
	docker-compose --file compose-files/load-tests-sqlite.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: load-tests-mariadb
load-tests-mariadb:
	docker-compose --file compose-files/load-tests-mariadb.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

## Docker things

.PHONY: server-docker-image
server-docker-image: wire
	docker build --tag $(SERVER_DOCKER_IMAGE_NAME):latest --file dockerfiles/server.Dockerfile .

.PHONY: push-server-to-docker
push-server-to-docker: prod-server-docker-image
	docker push $(SERVER_DOCKER_REPO_NAME):latest

## Running

.PHONY: dev
dev:
	docker-compose --file compose-files/development.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: run
run:
	docker-compose --file compose-files/production.json up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit
