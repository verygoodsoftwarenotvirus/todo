FROM golang:stretch

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .

#ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration" ]

# for a more specific test:
ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration", "-run", "TestAuditLogEntries/Listing/should_be_able_to_be_read_in_a_list_by_an_admin" ]
