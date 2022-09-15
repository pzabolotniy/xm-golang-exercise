lint:
	golangci-lint run

test:
	go test -cover -count 1 -gcflags "all=-l" ./...