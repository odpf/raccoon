.PHONY: all

ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")
APP_EXECUTABLE="out/clickstream-service"

setup:
	go mod tidy -v

source:
	source ".env.sample"

build-deps:
	go mod tidy -v

update-deps:
	go mod tidy -v

compile:
	mkdir -p out/
	go build -o $(APP_EXECUTABLE)

build: copy-config build-deps compile

install:
	go install $(ALL_PACKAGES)

fmt:
	go fmt $(ALL_PACKAGES)

vet:
	go vet $(ALL_PACKAGES)

lint:
	@for p in $(ALL_PACKAGES); do \
		echo "==> Linting $$p"; \
		golint $$p | { grep -vwE "exported (var|function|method|type|const) \S+ should have comment" || true; } \
	done

test:
	make lint
	ENVIRONMENT=test go test $(ALL_PACKAGES) -p=2 -v

test_ci:
	ENVIRONMENT=test go test $(ALL_PACKAGES) -p=1 -race

copy-config:
	cp application.yml.sample application.yml

start:
	./$(APP_EXECUTABLE) start

copy-config-ci:
	cp application.yml.ci application.yml

run:
	docker-compose build
	docker-compose up -d

ps:
	docker-compose ps

kill:
	docker-compose kill
