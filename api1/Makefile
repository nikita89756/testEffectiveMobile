BINARY_NAME=person
BINARY_PATH=./bin/$(BINARY_NAME)
CMD_PATH=./cmd/mainapi/main.go
MIGRATIONS_DIR=./internal/storage/postgres/migrations
DATABASE = postgres
DSN = "host=localhost user=nikita password=password123 dbname=persondb port=5432 sslmode=disable"
DOCKER_IMAGE_NAME=person
DOCKER_TAG=latest
DOCKER_IMAGE_FULL=$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

run: swag build 
	$(BINARY_PATH)

build: 
	@mkdir -p ./bin 
	@go build -o $(BINARY_PATH) $(CMD_PATH)

tests: 
	go test -v ./...

swag: 
	swag init -g $(CMD_PATH)
migrate-up: 
	goose -dir $(MIGRATIONS_DIR) $(DATABASE) $(DSN) up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(DATABASE) $(DSN) down

docker-compose-up: 
	docker-compose up -d --build

docker-compose-down:
	docker-compose down