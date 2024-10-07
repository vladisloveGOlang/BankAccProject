GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
 
ifneq (,$(wildcard ./.env))
    include .env
	export $(shell sed 's/=.*//' .env)
endif
 
.PHONY: help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@echo "  ${YELLOW}help            ${RESET} Show this help message"
	@echo "  ${YELLOW}build           ${RESET} Build application binary"
	@echo "  ${YELLOW}setup           ${RESET} Setup local environment"
	@echo "  ${YELLOW}check           ${RESET} Run tests, linters and tidy of the project"
	@echo "  ${YELLOW}test            ${RESET} Run tests only"
	@echo "  ${YELLOW}lint            ${RESET} Run linters via golangci-lint"
	@echo "  ${YELLOW}tidy            ${RESET} Run tidy for go module to remove unused dependencies"
	@echo "  ${YELLOW}run-web         ${RESET} Run web application example"
	@echo "  ${YELLOW}run-cli         ${RESET} Run cli application example"
	@echo "  ${YELLOW}gen        	  ${RESET} Generate openapi"

.PHONY: build
build: 
	OS="$(OS)" APP="web" ./hacks/build.sh
	OS="$(OS)" APP="cli" ./hacks/build.sh

.PHONY: skaffold
skaffold: 
	skaffold dev
	
.PHONY: build-docker
build-docker: 
	sudo chmod 666 /var/run/docker.sock
	docker build -t krisch/crm-backend:latest .

.PHONY: build-docker-x64
build-docker-x64: 
	sudo chmod 666 /var/run/docker.sock
	docker build --platform linux/x86_64 -t krisch/crm-backend:latest .

.PHONY: image-check-version
image-check-version: 
	sudo chmod 666 /var/run/docker.sock
	docker run krisch/crm-backend:latest cli
 
.PHONY: setup
setup:
	cp .env.example .env

.PHONY: check
check: %: gen tidy lint test

.PHONY: test
test:
	TEST_RUN_ARGS="$(TEST_RUN_ARGS)" TEST_DIR="$(TEST_DIR)" ./hacks/run-tests.sh

.PHONY: lint
lint:
	golangci-lint run --out-format=colored-line-number

.PHONY: fix
fix:
	golangci-lint run --fix  --out-format=colored-line-number

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: upgrade
upgrade:
	go get -u ./... && go mod tidy -v

.PHONY: run-web
run-web:
	go run ./cmd/web/.

.PHONY: run-web-docker
run-web-docker:
	 
	docker run --env-file=./.env  -p 8080:8080 krisch/crm-backend:latest web

.PHONY: run-dev
run-dev:
	sudo chmod 666 /var/run/docker.sock
	cd ./supertest && docker-compose run  

.PHONY: run-cli
run-cli:
	go run ./cmd/cli/.

.PHONY: github
github:
	sudo chmod 666 /var/run/docker.sock
	act --remote-name main --job docker --secret-file=.env

.PHONY: gen
gen:  
	for tag in omain oprofile ofederation oproject otask oreminder ocatalog; do \
		rm -rf ./internal/web/$$tag/; \
		mkdir -p ./internal/web/$$tag; \
	done
	oapi-codegen -config openapi/.openapi  -include-tags federation -package ofederation openapi/openapi.yaml > ./internal/web/oBankAcc/api.gen.go
	oapi-codegen -config openapi/.openapi  -include-tags about,health -package omain  openapi/openapi.yaml > ./internal/web/omain/api.gen.go
	oapi-codegen -config openapi/.openapi  -include-tags profile -package oprofile  openapi/openapi.yaml > ./internal/web/oprofile/api.gen.go
	oapi-codegen -config openapi/.openapi  -include-tags federation -package ofederation  openapi/openapi.yaml > ./internal/web/ofederation/api.gen.go
	oapi-codegen -config openapi/.openapi  -include-tags project -package oproject  openapi/openapi.yaml > ./internal/web/oproject/api.gen.go
	oapi-codegen -config openapi/.openapi  -include-tags task -package otask openapi/openapi.yaml > ./internal/web/otask/api.gen.go
	oapi-codegen -config openapi/.openapi  -include-tags reminder -package oreminder openapi/openapi.yaml > ./internal/web/oreminder/api.gen.go
	oapi-codegen -config openapi/.openapi  -include-tags catalog -package ocatalog openapi/openapi.yaml > ./internal/web/ocatalog/api.gen.go

.PHONY: registry-init
registry-init:
	yc init
	yc iam key create --service-account-name github -o key.json
	cat key.json | docker login --username json_key --password-stdin   cr.yandex/crprhclm83hhnolm61f4

.PHONY: registry-push
registry-push:
	docker tag krisch/crm-backend:latest cr.yandex/crprhclm83hhnolm61f4/crm-back:latest
	docker push cr.yandex/crprhclm83hhnolm61f4/crm-back:latest

.PHONY: testify
testify: 
	cd testify && npm run test --testNamePattern "${FILTER}"

.PHONY: testify-pull
testify-pull:
	rm -rf ./testify
	git clone https://github.com/krisch/testify.git 
	cd testify && npm install
	
.PHONY: yc
yc-update:
	yc compute instance  --folder-id b1gf0moi66tq4rhkd7d4 list
	yc compute instance update-container --folder-id b1gf0moi66tq4rhkd7d4 --container-image=cr.yandex/crprhclm83hhnolm61f4/crm-back:main-ce19b3f crm-back-api 
	
	
.PHONY: migrate-new
migrate-new:
	migrate create -ext sql -dir ./migrations ${NAME}

.PHONY: migrate
migrate:
	migrate -path ./migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" up 

.PHONY: migrate-down
migrate-down:
	migrate -path ./migrations -database "postgres://default:secret@postgres:5432/main?sslmode=disable" down  
	
.PHONY: wire
wire:
	wire ./internal/app/