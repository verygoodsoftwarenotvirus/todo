GOPATH       := $(GOPATH)
ARTIFACTS_DIR := artifacts
COVERAGE_OUT := $(ARTIFACTS_DIR)/coverage.out

KUBERNETES_NAMESPACE     := todo
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

## Testing things

.PHONY: lint
lint:
	docker run \
		--interactive \
		--tty \
		--volume=`pwd`:/go/src/`pwd` \
		--workdir=/go/src/`pwd` \
		--env=GO111MODULE=on \
		golangci/golangci-lint golangci-lint run --config=.golangci.yml ./...

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
	set -ex; \
	echo "mode: set" > $(COVERAGE_OUT);
	for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|mock)'`; do \
		go test -coverprofile=profile.out -v -race -failfast $$pkg; \
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

.PHONY: frontend-tests
frontend-tests:
	docker-compose --file compose-files/frontend-tests.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: integration-tests
integration-tests:
	docker-compose --file compose-files/integration-tests-postgres.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

# this is just a handy lil' helper I use sometimes
.PHONY: lintegration-tests
lintegration-tests: integration-tests lint

.PHONY: load-tests
load-tests:
	docker-compose --file compose-files/load-tests.yaml up \
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
	docker-compose --file compose-files/integration-coverage.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit
	go tool cover -html=./artifacts/integration-coverage.out

## CI-specific tasks

.PHONY: ci-load-tests
ci-load-tests:
	docker-compose --file compose-files/ci-load-tests.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: integration-tests-postgres
integration-tests-postgres: integration-tests

.PHONY: integration-tests-sqlite
integration-tests-sqlite:
	docker-compose --file compose-files/integration-tests-sqlite.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: integration-tests-mariadb
integration-tests-mariadb:
	docker-compose --file compose-files/integration-tests-mariadb.yaml up \
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
	docker-compose --file compose-files/development.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

.PHONY: run
run:
	docker-compose --file compose-files/production.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

## Minikube

.PHONY: wipe-minikube
wipe-minikube:
	minikube delete || true
	minikube start --vm-driver=virtualbox \
		--extra-config=apiserver.Authorization.Mode=RBAC
	sleep 10
	kubectl create clusterrolebinding ks-default --clusterrole cluster-admin --serviceaccount=kube-system:default

.PHONY: install-helm
install-helm:
	kubectl create serviceaccount tiller --namespace kube-system
	kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount kube-system:tiller
	helm init --upgrade --service-account tiller

.PHONY: init-local-cluster
init-local-cluster: install-helm install-chart

.PHONY: install-chart
install-chart:
	helm upgrade todo ./deploy/helm \
		--install \
		--force \
		--debug \
		--namespace=$(KUBERNETES_NAMESPACE)
