FROM golang:stretch AS compile-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev wget curl unzip

ADD . .

RUN go test -v -failfast -c -o /test gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/frontend

# FROM selenium/standalone-firefox:3.141.59-oxygen

# COPY --from=compile-stage /test /test

ENTRYPOINT [ "/test" ]
