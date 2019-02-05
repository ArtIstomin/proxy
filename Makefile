GOPATH=$(HOME)/go
BIN_DIR = $(GOPATH)/bin

install: 
	go mod download

test: lint
	go test ./...

lint:
	$(BIN_DIR)/golint

run:
	go run cmd/proxy/main.go

build-container:
	docker build -f build/proxy/Dockerfile --tag=myproxy .

run-container:
	docker run -p 80:80 -p 443:443 myproxy

run-container-with-build: build-container run-container
	