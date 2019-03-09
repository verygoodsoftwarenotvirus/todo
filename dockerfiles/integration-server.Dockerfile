# build stage
FROM golang:alpine AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update make git gcc musl-dev

ADD . .

RUN pwd
RUN ls -Al

RUN go build -o /todo -v gitlab.com/verygoodsoftwarenotvirus/todo/cmd/server/v1

# final stage
FROM alpine:latest

COPY config_files config_files
COPY --from=build-stage /todo /todo

EXPOSE 443 80

ENTRYPOINT ["/todo"]
