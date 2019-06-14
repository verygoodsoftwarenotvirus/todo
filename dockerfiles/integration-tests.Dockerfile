FROM golang:stretch

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

ADD . .

ENTRYPOINT [ "go", "test", "-v", "-failfast", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration" ]

# for a more specific test:
# ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/integration", "-run", "TestExport/Exporting/should_be_exportable" ]
