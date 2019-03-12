FROM python:3.7.2

ADD mypy.ini mypy.ini
RUN pip install --upgrade pip && pip install mypy

CMD [ "mypy", "--config-file", "./mypy.ini", "client/v1/python" ]
