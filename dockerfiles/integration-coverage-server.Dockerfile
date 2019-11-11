# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

ADD . .

RUN go test -o /todo -c -coverpkg \
	gitlab.com/verygoodsoftwarenotvirus/todo/internal/..., \
	gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/..., \
	gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/..., \
	gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1/ \
    gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# frontend-build-stage
FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend/v1 .

RUN npm install && npm run build

# final stage
FROM debian:stable

COPY config_files config_files
COPY --from=build-stage /todo /todo

EXPOSE 80

ENTRYPOINT ["/todo", "-test.coverprofile=/home/integration-coverage.out"]

