FROM alpine:latest

ADD . /code
WORKDIR /code

RUN sh setup.sh

RUN adduser node
USER node

CMD sh init.sh
