.PHONY: tests

fakery:
	@cd main && go build -o fakery main.go

tests: 
	@cd tests && go test ./... && go test --race ./...

check:
	@go vet ./...
