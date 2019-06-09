# build stage
FROM golang:stretch AS build-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

ADD . .

RUN go build -o /app gitlab.com/verygoodsoftwarenotvirus/todo/cmd/temp

ENTRYPOINT [ "/app" ]
