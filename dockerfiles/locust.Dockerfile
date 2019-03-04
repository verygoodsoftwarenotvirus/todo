# inspired by https://github.com/sernst/locusts/blob/682da15c6a53879c4af6875c6ce73899e82e9d6a/Dockerfile
FROM ubuntu:18.04

WORKDIR /opt

COPY client client
COPY tests/v1/load/locust .
COPY requirements.txt requirements.txt

RUN apt-get -y update && apt-get -y install \
    tree \
    libevent-dev \
    python3.6-dev \
    python3.6 \
    python3-pip && \
    pip3 install locustio && \
    pip3 install -r requirements.txt

EXPOSE 8089:8089
EXPOSE 5557:5557
EXPOSE 5558:5558

ENTRYPOINT [ "locust", "-f", "locustfile.py", "--host=http://todo-server" ]
