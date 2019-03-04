# https://github.com/sernst/locusts/blob/682da15c6a53879c4af6875c6ce73899e82e9d6a/Dockerfile
FROM ubuntu:18.04

WORKDIR /opt
COPY requirements.txt requirements.txt

COPY tests/v1/load/locust /scripts
COPY client /scripts/client

# COPY tests/v1/load/locust .
# COPY client client

COPY scripts/locust-entrypoint.py run.py
RUN echo '\n\
    {\n\
    "target": "http://todo-server",\n\
    "locusts": [\n\
    "TodoAPIServerLocust"\n\
    ]\n\
    }\n'\
    > /scripts/locust.config.json

RUN apt-get -y update && apt-get -y install \
    libevent-dev \
    python3.6-dev \
    python3.6 \
    python3-pip && \
    pip3 install -r requirements.txt

ENTRYPOINT [ "python3.6", "run.py", "--master-host=http://locust-leader" ]
# ENTRYPOINT [ "locust", "-f", '"/scripts/locustfile.py"', "--slave", "--host=http://todo-server", "--master-host=http://locust-leader", "TodoAPIServerLocust" ]
