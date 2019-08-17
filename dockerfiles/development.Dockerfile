# frontend-build-stage
FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend/v1 .

RUN npm install && npm run build

# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .
COPY --from=frontend-build-stage /app/public /frontend

RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# final stage
FROM debian:stretch

COPY --from=build-stage /todo /todo
COPY config_files config_files

RUN groupadd -g 999 appuser && \
    useradd -r -u 999 -g appuser appuser
USER appuser

ENV DOCKER=true

ENTRYPOINT ["/todo"]
