PWD                           := $(shell pwd)
GOPATH                        := $(GOPATH)
ARTIFACTS_DIR                 := artifacts
COVERAGE_OUT                  := $(ARTIFACTS_DIR)/coverage.out
SEARCH_INDICES_DIR            := $(ARTIFACTS_DIR)/search_indices
GO_FORMAT                     := gofmt -s -w
THIS                          := gitlab.com/verygoodsoftwarenotvirus/todo
PACKAGE_LIST                  := `go list $(THIS)/... | grep -Ev '(cmd|tests|testutil|mock|fake)'`
TEST_DOCKER_COMPOSE_FILES_DIR := environments/testing/compose_files
FRONTEND_DIR                  := frontend
FRONTEND_TOOL                 := pnpm
CONTAINER_RUNNER              := $(if $(shell command -v podman 2> /dev/null),podman,docker)

## non-PHONY folders/files

clear:
	@printf "\033[2J\033[3J\033[1;1H"

clean:
	rm -rf $(ARTIFACTS_DIR)

$(ARTIFACTS_DIR):
	@mkdir --parents $(ARTIFACTS_DIR)

clean-$(ARTIFACTS_DIR):
	@rm -rf $(ARTIFACTS_DIR)

$(SEARCH_INDICES_DIR):
	@mkdir --parents $(SEARCH_INDICES_DIR)

clean-search-indices:
	@rm -rf $(SEARCH_INDICES_DIR)

.PHONY: setup
setup: $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR) revendor frontend-vendor rewire configs

.PHONY: configs
configs:
	go run cmd/tools/config_gen/main.go

## prerequisites

ensure-wire:
ifndef $(shell command -v wire 2> /dev/null)
	$(shell GO111MODULE=off go get -u github.com/google/wire/cmd/wire)
endif

ensure-scc:
ifndef $(shell command -v scc 2> /dev/null)
	$(shell GO111MODULE=off go get -u github.com/boyter/scc)
endif

ensure-pnpm:
ifndef $(shell command -v pnpm 2> /dev/null)
	$(shell npm install -g pnpm)
endif

.PHONY: clean-vendor
clean-vendor:
	rm -rf vendor go.sum

vendor:
	if [ ! -f go.mod ]; then go mod init; fi
	go mod vendor

.PHONY: revendor
revendor: clean-vendor vendor frontend-vendor

## dependency injection

.PHONY: clean-wire
clean-wire:
	rm -f cmd/server/wire_gen.go

.PHONY: wire
wire: ensure-wire vendor
	wire gen $(THIS)/cmd/server

.PHONY: rewire
rewire: ensure-wire clean-wire wire

## Frontend stuff

.PHONY: clean-frontend
clean-frontend:
	@(cd $(FRONTEND_DIR) && rm -rf dist/build/)

.PHONY: frontend-vendor
frontend-vendor:
	@(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) install)

.PHONY: dev-frontend
dev-frontend: ensure-pnpm clean-frontend
	@(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run autobuild)

# frontend-only runs a simple static server that powers the frontend of the application. In this mode, all API calls are
# skipped, and data on the page is faked. This is useful for making changes that don't require running the entire service.
.PHONY: frontend-only
frontend-only: ensure-pnpm clean-frontend
	@(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run start:frontend-only)

## formatting

.PHONY: format-frontend
format-frontend:
	(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run format)

.PHONY: format-backend
format-backend:
	for file in `find $(PWD) -name '*.go'`; do $(GO_FORMAT) $$file; done

.PHONY: fmt
fmt: format

.PHONY: format
format: format-backend format-frontend

.PHONY: check-backend-formatting
check-backend-formatting: vendor
	$(CONTAINER_RUNNER) build --tag check-formatting:latest --file environments/testing/dockerfiles/formatting.Dockerfile .
	$(CONTAINER_RUNNER) run --rm check-formatting:latest

.PHONY: check-frontend-formatting
check-frontend-formatting:
	(cd $(FRONTEND_DIR) && $(FRONTEND_TOOL) run format:check)

.PHONY: check-formatting
check-formatting: check-backend-formatting check-frontend-formatting

## Testing things

.PHONY: docker-lint
docker-lint:
	$(CONTAINER_RUNNER) run --rm --volume `pwd`:`pwd` --workdir=`pwd` openpolicyagent/conftest:v0.21.0 test --policy docker_security.rego `find . -type f -name "*.Dockerfile"`

.PHONY: lint
lint:
	@$(CONTAINER_RUNNER) pull golangci/golangci-lint:latest
	$(CONTAINER_RUNNER) run \
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

.PHONY: quicktest # basically only running once instead of with -count 5 or whatever
quicktest: $(ARTIFACTS_DIR) vendor clear
	go test -cover -race -failfast $(PACKAGE_LIST)

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
lintegration-tests: lint clear integration-tests

.PHONY: integration-tests
integration-tests: integration-tests-sqlite integration-tests-postgres integration-tests-mariadb

.PHONY: integration-tests-
integration-tests-%:
	docker-compose \
	--file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-tests-base.yaml \
	--file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-tests-$*.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps $(if $(filter y yes true plz sure yup yep yass,$(LET_HANG)),, --abort-on-container-exit)

.PHONY: integration-coverage
integration-coverage: clean-$(ARTIFACTS_DIR) $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR) configs
	@# big thanks to https://blog.cloudflare.com/go-coverage-with-external-tests/
	rm -f $(ARTIFACTS_DIR)/integration-coverage.out
	@mkdir --parents $(ARTIFACTS_DIR)
	docker-compose \
	--file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-tests-base.yaml \
	--file $(TEST_DOCKER_COMPOSE_FILES_DIR)/integration_tests/integration-coverage.yaml up \
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
	docker-compose \
	--file $(TEST_DOCKER_COMPOSE_FILES_DIR)/load_tests/load-tests-base.yaml \
	--file $(TEST_DOCKER_COMPOSE_FILES_DIR)/load_tests/load-tests-$*.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps $(if $(filter y yes true plz sure yup yep yass,$(LET_HANG)),, --abort-on-container-exit)

## Running

.PHONY: dev
dev: clean-$(ARTIFACTS_DIR) $(ARTIFACTS_DIR) $(SEARCH_INDICES_DIR)
	docker-compose --file environments/local/docker-compose.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps $(if $(filter y yes true plz sure yup yep yass,$(LET_HANG)),, --abort-on-container-exit)

.PHONY: dev-user
dev-user:
	go run $(THIS)/cmd/tools/data_scaffolder --url=http://localhost --count=1 --single-user-mode --debug

.PHONY: load-data-for-admin
load-data-for-admin:
	go run $(THIS)/cmd/tools/data_scaffolder --url=http://localhost --count=5 --debug

## misc

.PHONY: tree
tree:
	tree -d -I vendor

.PHONY: cloc
cloc: ensure-scc
	@scc --include-ext go --exclude-dir vendor
