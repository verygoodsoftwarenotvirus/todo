GOPATH       := $(GOPATH)
COVERAGE_OUT := coverage.out

KUBERNETES_NAMESPACE     := todo
SERVER_DOCKER_IMAGE_NAME := todo-server
SERVER_DOCKER_REPO_NAME  := docker.io/verygoodsoftwarenotvirus/$(SERVER_DOCKER_IMAGE_NAME)

.PHONY: clean
clean:
	rm -f $(COVERAGE_OUT)
	rm -f example.db

.PHONY: dockercide
dockercide:
	docker system prune --force --all --volumes

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

$(COVERAGE_OUT):
	echo "mode: set" > coverage.out;
	for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests)'`; do \
		go test -coverprofile=profile.out -v -count 5 $$pkg; \
		cat profile.out | grep -v "mode: atomic" >> coverage.out; \
	done
	rm -f profile.out

.PHONY: test
test:
	docker build --tag coverage-$(SERVER_DOCKER_IMAGE_NAME):latest --file dockerfiles/coverage.Dockerfile .
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` coverage-$(SERVER_DOCKER_IMAGE_NAME):latest

.PHONY: integration-tests
integration-tests:
	docker-compose --file compose-files/integration-tests.yaml up --always-recreate-deps --build --remove-orphans --abort-on-container-exit --force-recreate

.PHONY: debug-integration-tests
debug-integration-tests: wire
	docker-compose --file compose-files/debug-integration-tests.yaml up --always-recreate-deps --build --remove-orphans --force-recreate

.PHONY: load-tests
load-tests: wire
	docker-compose --file compose-files/load-tests.yaml up --always-recreate-deps --build --remove-orphans --abort-on-container-exit --force-recreate

.PHONY: integration-coverage
integration-coverage:
	# big thanks to https://blog.cloudflare.com/go-coverage-with-external-tests/
	rm -f ./artifacts/integration-coverage.out
	mkdir -p ./artifacts
	docker-compose --file compose-files/integration-coverage.yaml up --always-recreate-deps --build --remove-orphans --force-recreate
	go tool cover -html=./artifacts/integration-coverage.out

## Docker things

.PHONY: server-docker-image
server-docker-image: wire
	docker build --tag $(SERVER_DOCKER_IMAGE_NAME):latest --file dockerfiles/server.Dockerfile .

.PHONY: prod-server-docker-image
prod-server-docker-image: wire
	docker build --tag $(SERVER_DOCKER_REPO_NAME):latest --file dockerfiles/server.Dockerfile .

.PHONY: push-server-to-docker
push-server-to-docker: prod-server-docker-image
	docker push $(SERVER_DOCKER_REPO_NAME):latest

## Running

.PHONY: run
run: server-docker-image
	docker-compose --file compose-files/docker-compose.yaml up --always-recreate-deps --build --remove-orphans --abort-on-container-exit --force-recreate

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
