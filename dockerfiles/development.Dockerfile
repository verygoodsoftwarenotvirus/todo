# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .

RUN GOOS=js GOARCH=wasm go build -o /frontend/website.wasm gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/frontend
RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

ENV CONFIGURATION_FILEPATH=config_files/debug.toml
ENV DOCKER=true

ENTRYPOINT ["/todo"]
