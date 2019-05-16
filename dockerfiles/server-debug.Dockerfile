# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

ADD . .

RUN go build -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# final stage
FROM golang:stretch

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .
COPY --from=build-stage /todo /todo

ENV CONFIGURATION_FILEPATH=config_files/debug.toml
ENV DOCKER=true

RUN go build -o /frontend/website.wasm gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/frontend

ENTRYPOINT ["/todo"]
