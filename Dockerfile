FROM golang:alpine

RUN apk update && apk upgrade && apk add git

RUN mkdir $GOPATH/src/app

WORKDIR $GOPATH/src/app

COPY . .

RUN go get -u github.com/golang/dep/...

RUN dep ensure
