FROM golang:alpine

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update gcc musl-dev

ADD . .

ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration" ]
# ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration", "-run", "TestOAuth2Clients/Creating/should_be_creatable"] # for a more specific test
