# Description

This is a test task for REST service

# Build && Run

```bash
docker-compose up --build -d --remove-orphans
```

# Examples

## Create company
```bash
curl -vvv -s -X POST -d '{"name": "ltd", "code": "007", "country": "md", "website": "http://google.com", "phone": "+995987655443"}' http://localhost:8088/api/v1/companies
```

## Delete company
```bash
curl -vvv -s -X DELETE http://localhost:8088/api/v1/companies/ab030400-f554-495a-83a5-44c8d66be239
```
