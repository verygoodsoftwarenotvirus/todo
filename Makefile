PWD                      := $(shell pwd)
GOPATH                   := $(GOPATH)
ARTIFACTS_DIR            := artifacts
COVERAGE_OUT             := $(ARTIFACTS_DIR)/coverage.out
CONFIG_DIR               := config_files
DOCKER_GO                := docker run --interactive --tty --rm --volume `pwd`:`pwd` --user `whoami`:`whoami` --workdir=`pwd` golang:latest go
GO_FORMAT                := gofmt -s -w
PACKAGE_LIST             := `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|mock|fake)'`
DOCKER_FILES_DIR         := dockerfiles
DOCKER_COMPOSE_FILES_DIR := compose_files
SERVER_DOCKER_IMAGE_NAME := todo-server
SERVER_DOCKER_REPO_NAME  := docker.io/verygoodsoftwarenotvirus/$(SERVER_DOCKER_IMAGE_NAME)

$(ARTIFACTS_DIR):
	@mkdir -p $(ARTIFACTS_DIR)

## Go-specific prerequisite stuff

ensure-wire:
ifndef $(shell command -v wire 2> /dev/null)
	$(shell GO111MODULE=off go get -u github.com/google/wire/cmd/wire)
endif

ensure-go-junit-report:
ifndef $(shell command -v go-junit-report 2> /dev/null)
	$(shell GO111MODULE=off go get -u github.com/jstemmer/go-junit-report)
endif

.PHONY: dev-tools
dev-tools: ensure-wire ensure-go-junit-report

.PHONY: vendor-clean
vendor-clean:
	rm -rf vendor go.sum

.PHONY: vendor
vendor:
	if [ ! -f go.mod ]; then go mod init; fi
	go mod vendor

.PHONY: revendor
revendor: vendor-clean vendor

## dependency injection

.PHONY: wire-clean
wire-clean:
	rm -f cmd/server/v1/wire_gen.go

.PHONY: wire
wire: ensure-wire
	wire gen gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

.PHONY: rewire
rewire: ensure-wire wire-clean wire

## Config

clean-configs:
	rm -rf $(CONFIG_DIR)

$(CONFIG_DIR):
	@mkdir -p $(CONFIG_DIR)
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

.PHONY: clean-coverage
clean-coverage:
	@rm -f $(COVERAGE_OUT) profile.out;

.PHONY: coverage
coverage: clean-coverage $(ARTIFACTS_DIR)
	@go test -coverprofile=$(COVERAGE_OUT) -covermode=atomic -race $(PACKAGE_LIST) > /dev/null
	@go tool cover -func=$(ARTIFACTS_DIR)/coverage.out | grep 'total:' | xargs | awk '{ print "COVERAGE: " $$3 }'

gitlab-ci-junit-report: $(ARTIFACTS_DIR) ensure-go-junit-report
	@mkdir $(CI_PROJECT_DIR)/test_artifacts
	go test -v -race -count 5 $(PACKAGE_LIST) | go-junit-report > $(CI_PROJECT_DIR)/test_artifacts/unit_test_report.xml

.PHONY: quicktest # basically only running once instead of with -count 5 or whatever
quicktest: $(ARTIFACTS_DIR)
	go test -cover -race -failfast $(PACKAGE_LIST)

.PHONY: format
format:
	for file in `find $(PWD) -name '*.go'`; do $(GO_FORMAT) $$file; done

.PHONY: check_formatting
check_formatting:
	docker build --tag check_formatting:latest --file $(DOCKER_FILES_DIR)/formatting.Dockerfile .
	docker run check_formatting:latest

.PHONY: frontend-tests
frontend-tests:
	docker-compose --file $(DOCKER_COMPOSE_FILES_DIR)/frontend-tests.yaml up \
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

.PHONY: integration-tests-
integration-tests-%:
	docker-compose --file $(DOCKER_COMPOSE_FILES_DIR)/integration-tests-$*.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: integration-coverage
integration-coverage: $(ARTIFACTS_DIR)
	@# big thanks to https://blog.cloudflare.com/go-coverage-with-external-tests/
	rm -f $(ARTIFACTS_DIR)/integration-coverage.out
	@mkdir -p $(ARTIFACTS_DIR)
	docker-compose --file $(DOCKER_COMPOSE_FILES_DIR)/integration-coverage.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit
	go tool cover -html=$(ARTIFACTS_DIR)/integration-coverage.out

## Load tests

.PHONY: load-tests
load-tests: load-tests-postgres load-tests-sqlite load-tests-mariadb

.PHONY: load-tests-
load-tests-%:
	docker-compose --file $(DOCKER_COMPOSE_FILES_DIR)/load-tests-$*.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

## Docker things

.PHONY: server-docker-image
server-docker-image: wire
	docker build --tag $(SERVER_DOCKER_IMAGE_NAME):latest --file $(DOCKER_FILES_DIR)/server.Dockerfile .

.PHONY: push-server-to-docker
push-server-to-docker: server-docker-image
	docker push $(SERVER_DOCKER_REPO_NAME):latest

## Running

.PHONY: dev
dev:
	docker-compose --file $(DOCKER_COMPOSE_FILES_DIR)/development.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: run
run:
	docker-compose --file $(DOCKER_COMPOSE_FILES_DIR)/production.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit
