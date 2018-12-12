# db build stage
FROM golang:alpine AS db-build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update gcc musl-dev

ADD . .

RUN go run tests/integration/db_bootstrap/main.go /example.db

# build stage
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
COPY --from=db-build-stage /example.db example.db

EXPOSE 443

ENTRYPOINT ["/todo"]
