# build stage
FROM golang:1.17-bullseye AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

# we need the `-tags json1` so sqlite can support JSON columns.
RUN go build -tags json1 -trimpath -o /todo -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# final stage
FROM debian:stretch

COPY --from=build-stage /todo /todo

ENTRYPOINT ["/todo"]
