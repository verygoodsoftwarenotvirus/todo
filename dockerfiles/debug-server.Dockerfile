# build stage
FROM golang:alpine AS db-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update gcc musl-dev

ADD . .

RUN go run gitlab.com/verygoodsoftwarenotvirus/todo/tools/db_bootstrap/v1/main.go

# server build stage
FROM golang:alpine AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update gcc musl-dev

ADD . .

RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# final stage
FROM alpine:latest

COPY dev_files/certs/server /certs
COPY database database
COPY --from=build-stage /todo /todo
COPY --from=db-stage /go/src/gitlab.com/verygoodsoftwarenotvirus/todo/example.db example.db

EXPOSE 443

ENTRYPOINT ["/todo"]
