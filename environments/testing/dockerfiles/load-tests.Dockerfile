# build stage
FROM golang:1.17-stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

RUN go build -o /loadtester gitlab.com/verygoodsoftwarenotvirus/todo/tests/load

# final stage
FROM debian:bullseye

COPY --from=build-stage /loadtester /loadtester

ENV DOCKER=true

ENTRYPOINT ["/loadtester"]
