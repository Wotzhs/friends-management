FROM golang:alpine

RUN apk update && apk upgrade && apk add git

RUN mkdir $GOPATH/src/app

WORKDIR $GOPATH/src/app

RUN go get -u -v github.com/lib/pq

RUN go get -u -v github.com/golang-migrate/migrate