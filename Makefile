build:
	go build -o run .
	
lint: 
	golangci-lint run ./...
	
ut:
	go test -v -count=1 -gcflags=-l -timeout=200s ./...

.PHONY: build lint ut
