FROM golang:alpine

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update gcc musl-dev

ADD . .

ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration/go" ]
# # for a more specific test:
# ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration/go", "-run", "TestAuth/should_accept_a_login_cookie_if_a_token_is_missing" ]
