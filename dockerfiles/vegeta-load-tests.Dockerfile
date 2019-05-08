# build stage
FROM golang:alpine AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update make git gcc musl-dev

ADD . .

RUN go build -o /loadtester gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/load/vegeta

# final stage
FROM alpine:latest

COPY --from=build-stage /loadtester /loadtester

ENV DOCKER=true

ENTRYPOINT ["/loadtester"]
