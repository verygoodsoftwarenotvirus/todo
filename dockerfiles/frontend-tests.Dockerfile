FROM golang:stretch AS compile-stage

WORKDIR /go/src/gitlab.com/verygoodsoftwarenotvirus/todo

RUN apt-get update -y && apt-get install -y make git gcc musl-dev wget curl unzip

ADD . .

RUN go test -v -failfast -c -o /test gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/frontend

FROM selenium/standalone-firefox:3.141.59-oxygen

COPY --from=compile-stage /test /test

# RUN wget https://chromedriver.storage.googleapis.com/$(curl https://chromedriver.storage.googleapis.com/LATEST_RELEASE)/chromedriver_linux64.zip -O chromedriver.zip

# RUN unzip chromedriver.zip
# RUN rm -f chromedriver.zip

# RUN chmod +x chromedriver

# RUN cp chromedriver /usr/bin/chromedriver

ENTRYPOINT [ "/test" ]
