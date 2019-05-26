# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

ADD . .

RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# frontend-build-stage
FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend .

RUN npm install && npm run build

# final stage
FROM debian:stable

COPY config_files config_files
COPY --from=build-stage /todo /todo
COPY --from=frontend-build-stage /app/dist /frontend

ENV CONFIGURATION_FILEPATH=config_files/production.toml

ENV DOCKER=true
EXPOSE 443 80

ENTRYPOINT ["/todo"]
