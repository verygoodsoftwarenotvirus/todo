FROM golang:alpine

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apk add --update make git gcc musl-dev

ADD . .

ADD ./scripts/coverage.sh /coverage.sh

RUN echo "mode: set" > coverage.out;

ENTRYPOINT 	for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|tools)'`; do \
		go test -coverprofile=profile.out -v -count 5 $$pkg; \
		cat profile.out | grep -v "mode: atomic" >> coverage.out; \
	done

