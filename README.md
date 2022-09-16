# Description

This is a test task for REST service

# Linter
Install [golangci-lint](https://golangci-lint.run/usage/install/)

```bash
$ cd $PROJECT_ROOT
$ make lint
```

# Run tests

```bash
$ cd $PROJECT_ROOT
$ make test
```

# Build && Run

```bash
docker-compose up --build -d --remove-orphans
```

## Known problems
if api didn't start, just restart it
```bash
docker-compose up -d
```

# Examples

## Generate JWT
```bash
docker exec -it xm-golang-exercise-api-1 /go/bin/tokengen
```

## Create company
```bash
curl -vvv -s -X POST \
  -H 'Authorization: Bearer **TOKEN**' \
  -d '{"name": "ltd", "code": "007", "country": "md", "website": "http://google.com", "phone": "+995987655443"}' \
  http://localhost:8088/api/v1/companies
```

## Delete company
```bash
curl -vvv -s -X DELETE \
  -H 'Authorization: Bearer **TOKEN**' \
  http://localhost:8088/api/v1/companies/ab030400-f554-495a-83a5-44c8d66be239
```

## Get company
```bash
curl -vvv -s http://localhost:8088/api/v1/companies/3997db3d-f747-4f00-adf8-1d2c71d2a911
```

## Get list of companies (using search)
```bash
curl -vvv -s -X POST \
  -d '{"companies_ids":["03da6341-950d-48a5-978c-9d53f155806a", "2025f015-e548-4620-a599-ff7ed3221b4f", "foo"]}' \
  http://localhost:8088/api/v1/search/companies
```