GOPATH                    := $(GOPATH)
PKG_ROOT                  := gitlab.com/verygoodsoftwarenotvirus/newsman
PKG                       := $(PKG_ROOT)
ARTIFACTS_DIR             := artifacts
COVERAGE_PATH             := $(ARTIFACTS_DIR)/coverage.out
UNIT_COVERAGE_PATH        := $(ARTIFACTS_DIR)/unit-coverage.out
INTEGRATION_COVERAGE_PATH := $(ARTIFACTS_DIR)/integration-coverage.out

## Go-specific prerequisite stuff
.PHONY: vendor-clean
vendor-clean:
	rm -rf vendor go.sum

vendor:
	GO111MODULE=on go mod vendor

.PHONY: revendor
revendor: vendor-clean vendor

.PHONY: test
test:
	set -ex; go test -v -cover $(PKG)

# Coverage

.PHONY: clean-unit-coverage
clean-unit-coverage:
	rm -f $(UNIT_COVERAGE_PATH)

$(ARTIFACTS_DIR):
	mkdir -p $(ARTIFACTS_DIR)

.PHONY: clean-integration-coverage
clean-integration-coverage:
	rm -f $(INTEGRATION_COVERAGE_PATH)

.PHONY: unit-coverage
unit-coverage: clean-unit-coverage $(ARTIFACTS_DIR)
	# note: we are deliberately not using race detection because we need the covermodes of both profiles to line up
	# run this test in conjunction with the regular test target (which does include race detection)
	go test -coverprofile=$(UNIT_COVERAGE_PATH) $(PKG)

.PHONY: integration-coverage
integration-coverage: clean-integration-coverage $(ARTIFACTS_DIR)
	docker-compose --file coverage.yaml up --always-recreate-deps --build --remove-orphans --force-recreate --abort-on-container-exit

coverage: unit-coverage integration-coverage
	go run cmd/tools/gocovmerge/gocovmerge.go $(UNIT_COVERAGE_PATH) $(INTEGRATION_COVERAGE_PATH) > $(COVERAGE_PATH)

# Docker things

.PHONY: docker-image
docker-image:
	docker build --tag 'newsman:latest' .

.PHONY: run-example
run-example:
	docker-compose --file example.yaml up --always-recreate-deps --build --remove-orphans --force-recreate --abort-on-container-exit

.PHONY: add-version
add-version:
ifdef branch
	git fetch
	git tag v$(branch) $(branch)
	git push origin :refs/tags/$(branch)
	git push --tags;
else
	@echo "branch undefined"
endif
