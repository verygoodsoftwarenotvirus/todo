FROM python:3.7.2

RUN pip install --upgrade pip && pip install mypy

CMD [ "mypy", "--config-file", "./mypy.ini", "client/v1/python" ]
