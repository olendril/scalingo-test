FROM golang:1.23
LABEL maintainer="Infrastructure Services Team <team-infrastructure-services@scalingo.com>"

RUN go install github.com/cespare/reflex@latest

WORKDIR $GOPATH/src/github.com/olendril/scalingo-test

EXPOSE 5000

CMD $GOPATH/bin/scalingo-test