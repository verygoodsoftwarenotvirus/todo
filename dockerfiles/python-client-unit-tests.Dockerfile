FROM python:latest

COPY requirements.txt /requirements.txt
COPY client/v1/python .

RUN pip3 install -r /requirements.txt

CMD [ "nosetests", "tests" ]
