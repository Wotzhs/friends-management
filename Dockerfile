FROM golang:alpine

RUN apk update && apk upgrade && apk add git && apk add curl

RUN mkdir $GOPATH/src/app

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v3.3.1/migrate.linux-amd64.tar.gz | tar xvz

RUN mv migrate.linux-amd64 /bin/migrate

WORKDIR $GOPATH/src/app

RUN go get -u -v github.com/lib/pq

RUN go get -u -v github.com/golang-migrate/migrate

RUN curl -o wait-for https://raw.githubusercontent.com/eficode/wait-for/master/wait-for

CMD ["chmod", "+x", "wait-for"]

CMD ["chmod", "+x", "start.sh"]
