PWD                           := $(shell pwd)
GOPATH                        := $(GOPATH)
ARTIFACTS_DIR                 := artifacts
COVERAGE_OUT                  := $(ARTIFACTS_DIR)/coverage.out
SEARCH_INDICES_DIR            := $(ARTIFACTS_DIR)/search_indices
DOCKER_GO                     := docker run --interactive --tty --rm --volume $(PWD):$(PWD) --user `whoami`:`whoami` --workdir=$(PWD) golang:latest go
GO_FORMAT                     := gofmt -s -w
THIS                          := gitlab.com/verygoodsoftwarenotvirus/todo
PACKAGE_LIST                  := `go list $(THIS)/... | grep -Ev '(cmd|tests|testutil|mock|fake)'`
TEST_DOCKER_COMPOSE_FILES_DIR := environments/testing/compose_files

## non-PHONY folders/files

clean:
	rm -rf $(ARTIFACTS_DIR)

$(ARTIFACTS_DIR):
	@mkdir -p $(ARTIFACTS_DIR)

clean_$(ARTIFACTS_DIR):
	@rm -rf $(ARTIFACTS_DIR)

$(SEARCH_INDICES_DIR):
	@mkdir -p $(SEARCH_INDICES_DIR)

clean_search_indices:
	@rm -rf $(SEARCH_INDICES_DIR)

.PHONY: setup
setup: $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR) revendor frontend-vendor rewire config_files

.PHONY: config_files
config_files:
	go run cmd/tools/config_gen/main.go

## Go-specific prerequisite stuff

ensure-wire:
ifndef $(shell command -v wire 2> /dev/null)
	$(shell GO111MODULE=off go get -u github.com/google/wire/cmd/wire)
endif

.PHONY: dev-tools
dev-tools: ensure-wire

.PHONY: clean_vendor
clean_vendor:
	rm -rf vendor go.sum

vendor:
	if [ ! -f go.mod ]; then go mod init; fi
	go mod vendor

.PHONY: revendor
revendor: clean_vendor vendor frontend-vendor

## Frontend stuff

.PHONY: frontend-vendor
frontend-vendor:
	@(cd frontend/ && npm install)

.PHONY: format-frontend
format-frontend:
	@(cd frontend/ && npm run format)

## dependency injection

.PHONY: clean_wire
clean_wire:
	rm -f cmd/server/wire_gen.go

.PHONY: wire
wire: ensure-wire vendor
	wire gen $(THIS)/cmd/server

.PHONY: rewire
rewire: ensure-wire clean_wire wire

## Testing things

.PHONY: docker-security-lint
docker-security-lint:
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` openpolicyagent/conftest:latest test --policy docker_security.rego `find . -type f -name "*.Dockerfile"`

.PHONY: lint
lint:
	@docker pull golangci/golangci-lint:latest
	docker run \
		--rm \
		--volume `pwd`:`pwd` \
		--workdir=`pwd` \
		--env=GO111MODULE=on \
		golangci/golangci-lint:latest golangci-lint run --config=.golangci.yml ./...

.PHONY: clean_coverage
clean_coverage:
	@rm -f $(COVERAGE_OUT) profile.out;

.PHONY: coverage
coverage: clean_coverage $(ARTIFACTS_DIR)
	@go test -coverprofile=$(COVERAGE_OUT) -covermode=atomic -race $(PACKAGE_LIST) > /dev/null
	@go tool cover -func=$(ARTIFACTS_DIR)/coverage.out | grep 'total:' | xargs | awk '{ print "COVERAGE: " $$3 }'

.PHONY: quicktest # basically only running once instead of with -count 5 or whatever
quicktest: $(ARTIFACTS_DIR) vendor
	go test -cover -race -failfast $(PACKAGE_LIST)

.PHONY: format
format:
	for file in `find $(PWD) -name '*.go'`; do $(GO_FORMAT) $$file; done

.PHONY: check_formatting
check_formatting: vendor
	docker build --tag check_formatting:latest --file environments/testing/dockerfiles/formatting.Dockerfile .
	docker run --rm check_formatting:latest
	(cd frontend/ && npm run format-check)

.PHONY: frontend-tests
frontend-tests:
	docker-compose --file $(TEST_DOCKER_COMPOSE_FILES_DIR)/frontend-tests.yaml up \
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
integration-tests: integration-tests-sqlite integration-tests-postgres integration-tests-mariadb

.PHONY: integration-tests-
integration-tests-%:
	docker-compose --file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-tests-$*.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps $(if $(filter y yes true plz sure yup yep yass,$(KEEP_RUNNING)),, --abort-on-container-exit)

.PHONY: integration-coverage
integration-coverage: clean_$(ARTIFACTS_DIR) $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR) config_files
	@# big thanks to https://blog.cloudflare.com/go-coverage-with-external-tests/
	rm -f $(ARTIFACTS_DIR)/integration-coverage.out
	@mkdir -p $(ARTIFACTS_DIR)
	docker-compose --file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-coverage.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit
	go tool cover -html=$(ARTIFACTS_DIR)/integration-coverage.out

## Load tests

.PHONY: load-tests
load-tests: load-tests-sqlite load-tests-postgres load-tests-mariadb

.PHONY: load-tests-
load-tests-%:
	docker-compose --file $(TEST_DOCKER_COMPOSE_FILES_DIR)/load_tests/load-tests-$*.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps $(if $(filter y yes true plz sure yup yep yass,$(KEEP_RUNNING)),, --abort-on-container-exit)

## Running

.PHONY: dev
dev: clean_$(ARTIFACTS_DIR) $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR) config_files
	docker-compose --file environments/local/docker-compose.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps $(if $(filter y yes true plz sure yup yep yass,$(KEEP_RUNNING)),, --abort-on-container-exit)

.PHONY: load-data
load-data:
	go run $(THIS)/cmd/tools/data_scaffolder --url=http://localhost --count=5 --debug

.PHONY: dev-frontend
dev-frontend:
	@(cd frontend && rm -rf dist/build/ && npm run autobuild)

# frontend-only runs a simple static server that powers the frontend of the application. In this mode, all API calls are
# skipped, and data on the page is faked. This is useful for making changes that don't require running the entire service.
.PHONY: frontend-only
frontend-only:
	@(cd frontend && rm -rf dist/build/ && npm run frontend-only)

## misc

.PHONY: tree
tree:
	tree -d -I vendor