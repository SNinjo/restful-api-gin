# restful-api-gin &middot; ![Coverage](https://img.shields.io/badge/Coverage-80%25-brightgreen)

A RESTful API server includes

* Framework: Gin
* OpenAPI: Swagger
* Database: MongoDB
* ORM: mgm
* Test: Testify
* Environment: Docker
* Deployment: Docker Compose

## Usage

### Build Database

```shell
docker-compose up -d mongo mongo-express
```

[Mongo Express](http://localhost:8081)

* username: root
* password: pass

### Develop

```shell
go run main.go
```

update the Swagger documentation content

```shell
swag init
```

[Swagger](http://localhost:8080/docs/index.html)

### Test

```shell
go test -coverprofile cover.out ./user/...
go tool cover -html cover.out
```

### Deploy

```shell
docker build . -t restful-api-gin
docker-compose up -d
```
