# Experimental Go + Docker Compose application
This project allows user to create friends, list friends, block other users, subscribe to users and list subsribers

## Getting Started
### Prequisites 
This project currently is set to work in dockerized environment only, so you will need to install Docker and Docker Compose to get it working

### Starting the application
In project root directory:
```shell 
docker-compose up -d
```

### Resetting the application database
In project root directory:
```shell 
docker-compose down -v
```

## Running test
In project root directory:
```shell 
docker-compose run test
```

## Built with
This project is created using *mostly* standard libraries including but not limitted to:

+ net/http
+ httprouter
+ database/sql
+ lib/pq

## Note
This is a experimental Go web API aimed to boost my understanding on how web server _really_ works, which it is often absracted away in many web frameworks.

## License
MIT