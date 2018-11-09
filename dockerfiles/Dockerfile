# build stage
FROM golang:alpine AS build-stage
WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

ADD . .
RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# final stage
FROM alpine:latest

COPY dev_files/certs/server /certs
COPY --from=build-stage /todo /todo

EXPOSE 443

ENTRYPOINT ["/todo"]
