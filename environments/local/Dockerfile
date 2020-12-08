# frontend-build-stage
FROM node:latest AS frontend-build-stage

WORKDIR /app

ADD frontend/ .

RUN npm install -g pnpm
RUN pnpm install
RUN pnpm run build

# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .
COPY --from=frontend-build-stage /app/dist /frontend

RUN go build -trimpath -o /todo gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server

# final stage
FROM debian:stretch

RUN mkdir /home/appuser
RUN groupadd --gid 1000 appuser && \
    useradd --system --uid 1000 --gid appuser appuser
RUN chown appuser /home/appuser
WORKDIR /home/appuser
USER appuser

COPY --from=build-stage /todo /todo

ENV DOCKER=true

ENTRYPOINT ["/todo"]
