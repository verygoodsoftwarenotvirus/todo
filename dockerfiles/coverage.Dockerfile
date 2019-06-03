FROM golang:stretch

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN which make

RUN go get -u github.com/axw/gocov/gocov
ADD . .

CMD make coverage.out
