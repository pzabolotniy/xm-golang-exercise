FROM golang:1.19 AS builder

WORKDIR /build
COPY . .

RUN go build -o /go/bin/api cmd/webapi/main.go
COPY config.yaml /go/bin

EXPOSE 8080

CMD ["/go/bin/api"]