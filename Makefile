build:
	go build -o run .
	
lint: 
	golangci-lint run ./...
	
ut:
	go test -v -count=1 -timeout=200s ./...

.PHONY: build lint ut
