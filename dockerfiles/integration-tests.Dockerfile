FROM golang:stretch

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

ADD . .

ENTRYPOINT [ "go", "test", "-v", "-failfast", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration" ]

# # for a more specific test:
# ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration", "-run", "TestItems/Listing/should_be_able_to_be_read_in_a_list" ]
