FROM golang:alpine

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update gcc musl-dev

ADD . .

RUN go build -o /loadtest cmd/load_test/main.go

ENTRYPOINT [ "/loadtest" ]
