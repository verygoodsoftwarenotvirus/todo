# build stage
FROM golang:alpine AS build-stage
WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update gcc musl-dev

ADD . .
RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# final stage
FROM alpine:latest

COPY dev_files/certs/server /certs
COPY database database
COPY --from=build-stage /todo /todo

EXPOSE 443

ENTRYPOINT ["/todo"]
