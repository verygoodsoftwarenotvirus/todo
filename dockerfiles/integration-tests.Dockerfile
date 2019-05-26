FROM golang:stretch

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

ADD . .

ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration" ]

# # for a more specific test:
# ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration", "-run", "TestAuth/should_be_able_to_change_2FA_Token" ]
