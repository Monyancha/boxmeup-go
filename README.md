# Boxmeup Server

> This is a WIP [Go](https://go-lang.org) implementation of [Boxmeup](https://boxmeupapp.com).

[![](https://drone.chris-saylor.com/api/badges/cjsaylor/boxmeup-go/status.svg)](https://drone.chris-saylor.com/cjsaylor/boxmeup-go)

Boxmeup is a web and mobile application to help users keep track of what they have in their containers and how to find items in specific containers.

## Requirements

* [Go >= 1.9.2](https://golang.org) - For local development
* [Docker 17.05.0-ce+](https://www.docker.com) - For building and running in docker containers

## Setup

```bash
cp docker-compose-dev.yml docker-compose.yml
```

> Note: there is no production compse file yet.

Modify the docker compose file to suit needs.

```bash
docker-compose up -d
```

Bring your own mysql:
```bash
docker run -p 8080:8080 -e MYSQL_DSN=username:password@host:port/database cjsaylor/boxmeup-go
```

See `.env.sample` for available configurations.

## Testing

A Drone CI instance is used to run tests. To run them locally, setup Drone CLI (instructions for MacOS):

### Install Drone CLI

```bash
curl -L https://github.com/drone/drone-cli/releases/download/v0.8.0/drone_darwin_amd64.tar.gz | tar zx
sudo cp drone /usr/local/bin
```

or with Homebrew:

```bash
brew tap drone/drone
brew install drone
```

See [Drone CLI documentation](http://docs.drone.io/cli-installation/)

### Execute Tests

```bash
drone exec
```

## Development

In order to run tests you will need to prepare your MySQL db by running the [`schema.sql`](./schema.sql) on the MySQL db you plan to use. If you use the docker provided MySQL image specified in the `docker-compose.yml` file, you can run (on the running server):

```bash
# For local development
cat schema.sql | docker exec -i $(docker-compose ps -q mysql) mysql boxmeup -u boxmeup -pboxmeup
```

Create a user:

```bash
curl -X POST \
  http://localhost:8080/api/user/register \
  -H 'content-type: multipart/form-data' \
  -F email=test@test.com \
  -F password=test1234
```

Obtain a json-webtoken (for use in subsequent requests to the API):

```bash
curl -X POST \
  http://localhost:8080/api/user/login \
  -H 'content-type: multipart/form-data' \
  -F email=test@test.com \
  -F password=test1234
```

This will yield a token:

```bash
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MDQ5NTY5NjYsImlkIjoxLCJuYmYiOjE1MDQ1MjQ5NjYsInV1aWQiOiI5Yzk1MWIyNi05MGU1LTExZTctOTY0Ny0wMjQyYWMxMjAwMDIifQ.mgumlN4hQ5Wq3lmK1uiO9tAX21UOv7kLx5MYFI9KcdA"
}
```

Use the token in the header of further API requests (cURL example):

```bash
-H 'authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0OTU1NTI1MzgsImlkIjoyLCJuYmYiOjE0OTUxMjA1MzgsInV1aWQiOiJkMjU1MzY5OC0zYmRjLTExZTctYTU0NC0wODAwMjdkNGZkMjgifQ.ccSUP9AOrBplbwBs6e8dpTpePXHLipBSHvnYL1gFalw'
```

Dependencies are committed into the repo via `godeps`, so no `go install` required.

To build: `go build -o server ./bin`

To add a dependency:

* `go get godep` (If you don't already have it)
* `go get <pkg>`
* Use it somewhere in the code.
* `godep save`

## Proprietary Code

Proprietary code is included via hooks using the `go generate ./...` command. To include proprietary hooks, it must be included to `vendor/github.com/cjsaylor/boxmeup-hooks`.