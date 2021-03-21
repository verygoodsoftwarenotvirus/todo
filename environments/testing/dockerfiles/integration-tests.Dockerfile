FROM golang:stretch

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .

ENTRYPOINT [ "go", "test", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/integration" ]

# to debug a specific test:
# ENTRYPOINT [ "go", "test", "-parallel", "1", "-v", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/integration", "-run", "TestIntegration/TestSomething_*" ]
