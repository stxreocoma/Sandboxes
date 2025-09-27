.PHONY: fmt lint test ci

fmt:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test -v ./... -race -count=1

ci: fmt lint test