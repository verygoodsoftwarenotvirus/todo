FROM golang:stretch

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev

RUN groupadd -g 999 appuser && \
	useradd -r -u 999 -g appuser appuser
USER appuser

ADD . .

RUN echo "mode: set" > coverage.out;

ENTRYPOINT 	for pkg in `go list gitlab.com/verygoodsoftwarenotvirus/todo/... | grep -Ev '(cmd|tests|tools)'`; do \
	go test -coverprofile=profile.out -v -count 5 $$pkg; \
	cat profile.out | grep -v "mode: atomic" >> coverage.out; \
	done

