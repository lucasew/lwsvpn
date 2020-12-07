FROM alpine:latest

ADD . /code
WORKDIR /code

RUN sh setup.sh

CMD sh init.sh
