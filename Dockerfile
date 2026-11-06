FROM golang:alpine@sha256:ac09a5f469f307e5da71e766b0bd59c9c49ea460a528cc3e6686513d64a6f1fb as build

WORKDIR /
COPY ./go.mod .
COPY ./app.go .
RUN go build -o /app app.go

RUN echo "$(pwd; ls)"
FROM alpine:latest@sha256:3f4d6b32e26703c6082839234f43769d5046b3ac7c910033911d1e71d8cd630a
RUN apk add rclone

WORKDIR /code

RUN adduser node --disabled-password

COPY --chown="node:node" --from=build /app .

RUN echo "$(pwd; ls)"

ADD ./setup.sh .
RUN sh setup.sh

USER node

CMD /code/app
