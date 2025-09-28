.PHONY: fmt lint test ci

fmt:
	go fmt ./internal/...

lint:
	golangci-lint run ./internal/...

test:
	go test -v ./internal/... -race -count=1

ci: 
	@$(MAKE) fmt 
	@$(MAKE) lint 
	@$(MAKE) test