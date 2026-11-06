FROM golang:alpine@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 as build

WORKDIR /
COPY ./go.mod .
COPY ./app.go .
RUN go build -o /app app.go

RUN echo "$(pwd; ls)"
FROM alpine:latest@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412
RUN apk add rclone

WORKDIR /code

RUN adduser node --disabled-password

COPY --chown="node:node" --from=build /app .

RUN echo "$(pwd; ls)"

ADD ./setup.sh .
RUN sh setup.sh

USER node

CMD /code/app
