GOPATH       := $(GOPATH)
COVERAGE_OUT := coverage.out

KUBERNETES_NAMESPACE     := todo
SERVER_DOCKER_IMAGE_NAME := todo-server
SERVER_DOCKER_REPO_NAME  := docker.io/verygoodsoftwarenotvirus/$(SERVER_DOCKER_IMAGE_NAME)

SERVER_PRIV_KEY := dev_files/certs/server/key.pem
SERVER_CERT_KEY := dev_files/certs/server/cert.pem
CLIENT_PRIV_KEY := dev_files/certs/client/key.pem
CLIENT_CERT_KEY := dev_files/certs/client/cert.pem

.PHONY: clean
clean:
	rm -f $(COVERAGE_OUT)
	rm -f example.db

.PHONY: dockercide
dockercide:
	docker system prune --force --all --volumes

## Project prerequisites

.PHONY: dev-tools
dev-tools:
	go get -u github.com/google/wire/cmd/wire

.PHONY: wire-clean
wire-clean:
	rm -f cmd/server/v1/wire_gen.go

.PHONY: wire
wire:
	wire gen gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

.PHONY: rewire
rewire: wire-clean wire

.PHONY: prerequisites
prerequisites: vendor $(SERVER_PRIV_KEY) $(SERVER_CERT_KEY) $(CLIENT_PRIV_KEY) $(CLIENT_CERT_KEY)

$(SERVER_PRIV_KEY) $(SERVER_CERT_KEY):
	mkdir -p dev_files/certs/server
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=localhost" -keyout dev_files/certs/server/key.pem -out dev_files/certs/server/cert.pem

$(CLIENT_PRIV_KEY) $(CLIENT_CERT_KEY):
	mkdir -p dev_files/certs/client
	openssl req -newkey rsa:2048 -new -nodes -x509 -days 3650 -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=localhost" -keyout dev_files/certs/client/key.pem -out dev_files/certs/client/cert.pem

.PHONY: vendor
vendor:
	docker run --env GO111MODULE=on --volume `pwd`:`pwd` --workdir=`pwd` golang:latest /bin/sh -c "go mod vendor"
	sudo chown `whoami`:`whoami` vendor go.sum

.PHONY: revendor
revendor:
	rm -rf vendor go.sum
	docker run --env GO111MODULE=on --volume `pwd`:`pwd` --workdir=`pwd` golang:latest /bin/sh -c "go mod vendor"
	sudo chown `whoami`:`whoami` vendor go.sum

.PHONY: unvendor
unvendor:
	rm -rf vendor go.{mod,sum}
	GO111MODULE=on go mod init

## Testing things

$(COVERAGE_OUT):
	./scripts/coverage.sh

.PHONY: test
test:
	docker build --tag coverage-$(SERVER_DOCKER_IMAGE_NAME):latest --file dockerfiles/coverage.Dockerfile .
	docker run --rm --volume `pwd`:`pwd` --workdir=`pwd` coverage-$(SERVER_DOCKER_IMAGE_NAME):latest

.PHONY: integration-tests
integration-tests: wire
	docker-compose --file compose-files/integration-tests.yaml up --always-recreate-deps --build --remove-orphans --force-recreate --abort-on-container-exit

.PHONY: debug-integration-tests
debug-integration-tests: wire # literally the same except it won't exit
	docker-compose --file compose-files/integration-tests.yaml up --always-recreate-deps --build --remove-orphans --force-recreate

.PHONY: load-tests
load-tests: wire # literally the same except it won't exit
	docker-compose --file compose-files/load-tests.yaml up --always-recreate-deps --build --remove-orphans --force-recreate --abort-on-container-exit

## Docker things

.PHONY: server-docker-image
server-docker-image: prerequisites wire
	docker build --tag $(SERVER_DOCKER_IMAGE_NAME):latest --file dockerfiles/server.Dockerfile .


.PHONY: prod-server-docker-image
prod-server-docker-image: prerequisites wire
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

