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
FRONTEND_DIR                  := frontend
FRONTEND_TOOL                 := pnpm

## non-PHONY folders/files

clean:
	rm --recursive --force $(ARTIFACTS_DIR)

$(ARTIFACTS_DIR):
	@mkdir --parents $(ARTIFACTS_DIR)

clean_$(ARTIFACTS_DIR):
	@rm --recursive --force $(ARTIFACTS_DIR)

$(SEARCH_INDICES_DIR):
	@mkdir --parents $(SEARCH_INDICES_DIR)

clean_search_indices:
	@rm --recursive --force $(SEARCH_INDICES_DIR)

.PHONY: setup
setup: $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR) revendor frontend_vendor rewire config_files

.PHONY: config_files
config_files:
	go run cmd/tools/config_gen/main.go

## Go-specific prerequisite stuff

ensure_wire:
ifndef $(shell command -v wire 2> /dev/null)
	$(shell GO111MODULE=off go get -u github.com/google/wire/cmd/wire)
endif

ensure_pnpm:
ifndef $(shell command -v pnpm 2> /dev/null)
	$(shell npm install -g pnpm)
endif

.PHONY: clean_vendor
clean_vendor:
	rm --recursive --force vendor go.sum

vendor:
	if [ ! -f go.mod ]; then go mod init; fi
	go mod vendor

.PHONY: revendor
revendor: clean_vendor vendor frontend_vendor

## dependency injection

.PHONY: clean_wire
clean_wire:
	rm -f cmd/server/wire_gen.go

.PHONY: wire
wire: ensure_wire vendor
	wire gen $(THIS)/cmd/server

.PHONY: rewire
rewire: ensure_wire clean_wire wire

## Frontend stuff

.PHONY: clean_frontend
clean_frontend:
	@(cd $(FRONTEND_DIR) && rm --recursive --force dist/build/)

.PHONY: frontend_vendor
frontend_vendor:
	@(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) install)

.PHONY: dev_frontend
dev_frontend: ensure_pnpm clean_frontend
	@(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run autobuild)

# frontend_only runs a simple static server that powers the frontend of the application. In this mode, all API calls are
# skipped, and data on the page is faked. This is useful for making changes that don't require running the entire service.
.PHONY: frontend_only
frontend_only: ensure_pnpm clean_frontend
	@(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run frontend_only)

## formatting

.PHONY: format_frontend
format_frontend:
	(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run format)

.PHONY: format_backend
format_backend:
	for file in `find $(PWD) -name '*.go'`; do $(GO_FORMAT) $$file; done

.PHONY: format
format: format_backend format_frontend

.PHONY: check_backend_formatting
check_backend_formatting: vendor
	docker build --tag check_formatting:latest --file environments/testing/dockerfiles/formatting.Dockerfile .
	docker run --rm check_formatting:latest

.PHONY: check_frontend_formatting
check_frontend_formatting:
	(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run format_check)

.PHONY: check_formatting
check_formatting: check_backend_formatting check_frontend_formatting

## Testing things

.PHONY: docker_security_lint
docker_security_lint:
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

.PHONY: frontend_tests
frontend_tests:
	docker-compose --file $(TEST_DOCKER_COMPOSE_FILES_DIR)/frontend-tests.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

## Integration tests

.PHONY: lintegration_tests # this is just a handy lil' helper I use sometimes
lintegration_tests: integration_tests lint

.PHONY: integration_tests
integration_tests: integration_tests_sqlite integration_tests_postgres integration_tests_mariadb

.PHONY: integration_tests_
integration_tests_%:
	docker-compose --file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-tests-$*.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps $(if $(filter y yes true plz sure yup yep yass,$(KEEP_RUNNING)),, --abort-on-container-exit)

.PHONY: integration_coverage
integration_coverage: clean_$(ARTIFACTS_DIR) $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR) config_files
	@# big thanks to https://blog.cloudflare.com/go-coverage-with-external-tests/
	rm -f $(ARTIFACTS_DIR)/integration-coverage.out
	@mkdir --parents $(ARTIFACTS_DIR)
	docker-compose --file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-coverage.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit
	go tool cover -html=$(ARTIFACTS_DIR)/integration-coverage.out

## Load tests

.PHONY: load_tests
load_tests: load_tests_sqlite load_tests_postgres load_tests_mariadb

.PHONY: load_tests_
load_tests_%:
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

.PHONY: dev_user
dev_user:
	go run $(THIS)/cmd/tools/data_scaffolder --url=http://localhost --count=1 --single-user-mode --debug

.PHONY: load_data_for_admin
load_data_for_admin:
	go run $(THIS)/cmd/tools/data_scaffolder --url=http://localhost --count=5 --debug

## misc

.PHONY: tree
tree:
	tree -d -I vendor