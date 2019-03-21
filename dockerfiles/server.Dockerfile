# build stage
FROM golang:alpine AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update make git gcc musl-dev

ADD . .

RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# final stage
FROM alpine:latest

COPY config_files config_files
COPY --from=build-stage /todo /todo

ENV CONFIGURATION_FILEPATH=config_files/production.toml

ENV DOCKER=true
EXPOSE 443 80

ENTRYPOINT ["/todo"]
