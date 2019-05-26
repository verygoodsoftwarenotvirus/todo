GOPATH       := $(GOPATH)
ARTIFACTS_DIR := artifacts
INTEGRATION_COVERAGE_OUT := coverage.out

KUBERNETES_NAMESPACE     := todo
SERVER_DOCKER_IMAGE_NAME := todo-server
SERVER_DOCKER_REPO_NAME  := docker.io/verygoodsoftwarenotvirus/$(SERVER_DOCKER_IMAGE_NAME)

## dependency injectdion

.PHONY: install-dev-tool-wire
install-dev-tool-wire:
	go get -u github.com/google/wire/cmd/wire

.PHONY: wire-clean
wire-clean:
	rm -f cmd/server/v1/wire_gen.go

.PHONY: wire
wire:
	wire gen gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

.PHONY: rewire
rewire: wire-clean wire

## Go-specific prerequisite stuff
.PHONY: vendor-clean
vendor-clean:
	rm -rf vendor go.sum

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor

.PHONY: revendor
revendor: vendor-clean vendor

## Testing things

lint:
	GO111MODULE=on golangci-lint run --config=.golangci.yml ./...

$(INTEGRATION_COVERAGE_OUT):
	echo "mode: set" > $(INTEGRATION_COVERAGE_OUT);
	for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests)'`; do \
		go test -coverprofile=profile.out -v -count 5 $$pkg; \
		cat profile.out | grep -v "mode: atomic" >> $(INTEGRATION_COVERAGE_OUT); \
	done
	rm -f profile.out

.PHONY: test
test:
	docker build --tag coverage-$(SERVER_DOCKER_IMAGE_NAME):latest --file dockerfiles/coverage.Dockerfile .
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` coverage-$(SERVER_DOCKER_IMAGE_NAME):latest

.PHONY: integration-tests
integration-tests:
	docker-compose --file compose-files/integration-tests.yaml up \
	--build \
	--force-recreate \
	--remove-orphans \
	--renew-anon-volumes \
	--always-recreate-deps \
	--abort-on-container-exit

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

## Frontend things

.PHONY: frontend-dev
frontend-dev:
	docker build --tag frontend:latest --file=dockerfiles/frontend-dev.Dockerfile .
	docker run --publish 80 frontend:latest

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
	docker-compose --file compose-files/docker-compose.yaml up \
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
