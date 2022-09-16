FROM golang:1.19 AS builder

WORKDIR /build
COPY . .

RUN go build -o /go/bin/api cmd/webapi/main.go && \
    go build -o /go/bin/tokengen cmd/token/main.go
COPY config.yaml /go/bin

EXPOSE 8088

CMD ["/go/bin/api"]