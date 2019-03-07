# build stage
FROM golang:alpine AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update make git gcc musl-dev

ADD . .

ENV GO111MODULE=on
RUN go mod vendor

RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# final stage
FROM alpine:latest

COPY database database
COPY --from=build-stage /todo /todo

EXPOSE 443 80

ENTRYPOINT ["/todo"]
