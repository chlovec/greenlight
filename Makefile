# Load variables from .env file
include .env
export $(shell sed 's/=.*//' .env)

test:
	@echo "Environment: ${ENV}"
	@echo "SMTP Sender: ${SMTP_SENDER}"

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## lint: Run golangci-lint on the codebase
.PHONY: lint
lint:
	golangci-lint run

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## run/api/cors/simple: Run the simple CORS example
.PHONY:run/api/cors/simple
run/api/cors/simple:
	go run ./cmd/examples/cors/simple

## run/cors/preflight: Run the CORS preflight example
.PHONY: run/cors/preflight
run/cors/preflight:
	go run ./cmd/examples/cors/preflight

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=<file_name>: Create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migration name=<file_name>"; \
		exit 1; \
	fi
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}; \

## db/migrations/up: Apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@if [ -z "$(GREENLIGHT_DB_DSN)" ]; then \
		echo "Usage: GREENLIGHT_DB_DSN must be set or passed as a flag"; \
		exit 1; \
	fi
	@echo 'Running migrations up...'
	migrate -path=./migrations -database="$(GREENLIGHT_DB_DSN)" up

## db/migrations/down: Revert all database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@if [ -z "$(GREENLIGHT_DB_DSN)" ]; then \
		echo "Usage: GREENLIGHT_DB_DSN must be set or passed as a flag"; \
		exit 1; \
	fi
	@echo 'Running migrations down...'
	migrate -path=./migrations -database="$(GREENLIGHT_DB_DSN)" down

## db/migrations/force version=<version>: Force the database schema to a specific version
.PHONY: db/migrations/force
db/migrations/force: confirm
	@if [ -z "$(GREENLIGHT_DB_DSN)" ]; then \
		echo "Usage: GREENLIGHT_DB_DSN must be set or passed as a flag"; \
		exit 1; \
	fi
	@echo 'Running force migrations...'
	migrate -path=./migrations -database="$(GREENLIGHT_DB_DSN)" force ${version}
