# frontend-build-stage
FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend/v1 .

RUN npm install && npm run build

FROM golang:stretch

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .
COPY --from=frontend-build-stage /app/public /frontend

RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

ENV CONFIGURATION_FILEPATH=config_files/development.toml
ENV DOCKER=true

ENTRYPOINT ["/todo"]
