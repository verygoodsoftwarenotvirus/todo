# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

ADD . .

RUN go build -o /todo -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# frontend-build-stage
FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend/v1 .

RUN npm install && npm run build

# final stage
FROM debian:stretch

RUN groupadd -g 999 appuser && \
    useradd -r -u 999 -g appuser appuser
USER appuser

COPY config_files config_files
COPY --from=build-stage /todo /todo
COPY --from=frontend-build-stage /app/public /frontend

ENTRYPOINT ["/todo"]
