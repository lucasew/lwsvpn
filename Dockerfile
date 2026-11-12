FROM golang:alpine@sha256:eb61c3873c97416fff0116553fff65d67a3549f72e11d9758849a689cdd8b523 as build

WORKDIR /
COPY ./go.mod .
COPY ./app.go .
RUN go build -o /app app.go

RUN echo "$(pwd; ls)"
FROM alpine:latest@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf
RUN apk add rclone

WORKDIR /code

RUN adduser node --disabled-password

COPY --chown="node:node" --from=build /app .

RUN echo "$(pwd; ls)"

ADD ./setup.sh .
RUN sh setup.sh

USER node

CMD /code/app
