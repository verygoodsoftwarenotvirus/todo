# build stage
FROM golang:1.17-stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

# we need the `-tags json1` so sqlite can support JSON columns.
RUN go build -tags json1 -trimpath -o /workers -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/workers

# final stage
FROM debian:bullseye

COPY --from=build-stage /workers /workers

ENTRYPOINT ["/workers"]
