FROM golang:1.17-bullseye

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

COPY . .

ENTRYPOINT [ "go", "test", "-v", "-failfast", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/integration" ]

# to debug a specific test:
# ENTRYPOINT [ "go", "test", "-parallel", "1", "-v", "-failfast", "gitlab.com/verygoodsoftwarenotvirus/todo/tests/integration", "-run", "InsertTestNameHere" ]
