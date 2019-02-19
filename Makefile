GOPATH=$(HOME)/go
BIN_DIR = $(GOPATH)/bin
PROTO_DIR = ./internal/pkg/proto

install: 
	go mod download

test: lint
	go test ./...

lint:
	$(BIN_DIR)/golint

run:
	sudo go run cmd/proxy/main.go

build-container:
	docker build -f build/proxy/Dockerfile --tag=myproxy .
	docker build -f build/activity/Dockerfile --tag=activity .

run-compose:
	docker-compose -f deployments/docker-compose.yaml up

run-compose-with-build: build-container run-compose

run-k8s:
	./deployments/k8s/run-k8s.sh

proto-all: proto-activity

proto-activity:
	protoc -I=$(PROTO_DIR)/activity --go_out=plugins=grpc:$(PROTO_DIR)/activity $(PROTO_DIR)/activity/activity.proto