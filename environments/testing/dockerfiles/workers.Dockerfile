# build stage
FROM golang:1.17-stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

RUN go build -trimpath -o /workers -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/workers

# final stage
FROM debian:bullseye

COPY --from=build-stage /workers /workers

ENTRYPOINT ["/workers"]
