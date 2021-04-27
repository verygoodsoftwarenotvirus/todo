# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

COPY . .

# we need the `-tags json1` so sqlite can support JSON columns.
RUN go build -tags json1 -trimpath -o /todo -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# frontend-build-stage
FROM node:lts-stretch AS frontend-build-stage

WORKDIR /app

COPY frontend/ .

RUN npm install -g pnpm
RUN pnpm install
RUN pnpm run build

# final stage
FROM debian:stretch

COPY --from=build-stage /todo /todo
COPY --from=frontend-build-stage /app/dist /frontend

ENTRYPOINT ["/todo"]
