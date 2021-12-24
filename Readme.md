# Overview

Simple, extremely not production-ready toy messaging app in Go for the purposes of testing out frameworks and orms.

Currently built with Gin and Gorm on mysql.

# Development

## Setup

Set up and run containerized

```
make build-and-run
```

Tear down and clean up

```
make teardown
```

## Query locally

```
$ curl -d '{"username": "jolene"}' -H "Content-Type: application/json" -X POST localhost:8080/users
```