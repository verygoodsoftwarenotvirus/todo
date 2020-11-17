# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

RUN go build -trimpath -o /todo -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# frontend-build-stage
FROM node:lts-stretch AS frontend-build-stage

WORKDIR /app

COPY frontend/ .

RUN npm install && npm run build

# final stage
FROM debian:stretch

RUN mkdir /home/appuser
RUN groupadd --gid 999 appuser && \
    useradd --system --uid 999 --gid appuser appuser
RUN chown appuser /home/appuser
WORKDIR /home/appuser
USER appuser

COPY environments/testing/config_files/integration-tests-postgres.toml /etc/config.toml
COPY --from=build-stage /todo /todo
COPY --from=frontend-build-stage /app/dist /frontend

ENTRYPOINT ["/todo"]
