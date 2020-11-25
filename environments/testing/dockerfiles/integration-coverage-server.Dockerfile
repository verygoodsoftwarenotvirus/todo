# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

RUN go test -o /integration-server -c -coverpkg \
	gitlab.com/verygoodsoftwarenotvirus/todo/internal/..., \
	gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/ \
    gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# frontend-build-stage
FROM node:lts-stretch AS frontend-build-stage

WORKDIR /app

COPY frontend/ .

RUN npm install && npm run build

# final stage
FROM debian:stable

COPY --from=build-stage /integration-server /integration-server

EXPOSE 80

ENTRYPOINT ["/integration-server", "-test.coverprofile=/home/integration-coverage.out"]

