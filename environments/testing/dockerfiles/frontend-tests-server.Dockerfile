# frontend-build-stage
FROM node:lts-stretch AS frontend-build-stage

WORKDIR /app

COPY frontend/v2 .

RUN npm run install:clean

# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .
COPY --from=frontend-build-stage /app/dist /frontend

RUN go build -trimpath -o /todo -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# final stage
FROM debian:stretch

COPY --from=build-stage /todo /todo
COPY --from=frontend-build-stage /app/dist /frontend

RUN mkdir /home/appuser
RUN groupadd --gid 999 appuser && \
    useradd --system --uid 999 --gid appuser appuser
RUN chown appuser /home/appuser
WORKDIR /home/appuser
USER appuser

COPY environments/testing/config_files/frontend-tests.toml /etc/config.toml

ENTRYPOINT ["/todo"]
