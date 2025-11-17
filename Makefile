BIN_NAME=nats-client
ROOT_DIR=./
BUILD_DIR=$(ROOT_DIR)build
DEBIAN_PACKAGE_PATH=$(BUILD_DIR)/debian
BIN_DIR=$(BUILD_DIR)/bin
NFPM_CONFIG_DIR=$(ROOT_DIR)/debian/config

SHELL := /bin/bash

.PHONY: all clean

all: build


build: clean
	go build -v -o $(BUILD_DIR)/bin/$(BIN_NAME) -ldflags="-s -w -X main.version=$(CI_COMMIT_TAG)"  ./

build-portable: clean
	CGO_ENABLED=0 GOOS=linux go build -v -o $(BUILD_DIR)/bin/$(BIN_NAME)  ./

build-gc: clean
	go build -gcflags '-m -l' -v -o $(BUILD_DIR)/bin/$(BIN_NAME)  ./


docker-clean:
	docker image prune -f
	
docker-build: docker-clean
	docker build --tag mrtdeh/$(BIN_NAME) -f ./deploy/dockerfile .


docker-build-local: docker-clean build-portable
	docker build --tag mrtdeh/$(BIN_NAME) -f ./deploy/dockerfile.local .

docker-up-dc1:
	docker compose -p dc1 -f ./deploy/docker-compose-dc1.yml up --force-recreate --remove-orphans --build -d

docker-up:
	docker compose -p dc1 -f ./deploy/docker-compose-dc1.yml up --force-recreate --remove-orphans --build -d
	docker compose -p dc2 -f ./deploy/docker-compose-dc2.yml up --force-recreate --remove-orphans -d
	docker compose -p dc3 -f ./deploy/docker-compose-dc3.yml up --force-recreate --remove-orphans -d

docker-stop:
	docker compose -p dc1 -f ./deploy/docker-compose-dc1.yml stop 

docker-start:
	docker compose -p dc1 -f ./deploy/docker-compose-dc1.yml start 

docker-down:
	docker compose -p dc1 -f ./deploy/docker-compose-dc1.yml down 
	docker compose -p dc2 -f ./deploy/docker-compose-dc2.yml down 
	docker compose -p dc3 -f ./deploy/docker-compose-dc3.yml down 



clean:
	rm -rf $(BUILD_DIR)/*