# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

RUN go build -o /loadtester gitlab.com/verygoodsoftwarenotvirus/todo/tests/load

# final stage
FROM debian:stretch

COPY --from=build-stage /loadtester /loadtester

ENV DOCKER=true

ENTRYPOINT ["/loadtester"]
