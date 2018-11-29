FROM golang:alpine

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update make git gcc musl-dev

ADD . .
ADD ./scripts/coverage.sh /coverage.sh

ENTRYPOINT [ "/bin/sh", "/coverage.sh" ]
