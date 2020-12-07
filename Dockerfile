FROM golang:alpine as build

WORKDIR /
COPY ./go.mod .
COPY ./app.go .
RUN go build -o /app app.go

RUN echo "$(pwd; ls)"
FROM alpine:latest
RUN apk add rclone

WORKDIR /code

RUN adduser node --disabled-password

COPY --chown="node:node" --from=build /app .

RUN echo "$(pwd; ls)"

ADD ./setup.sh .
RUN sh setup.sh

USER node

CMD /code/app
