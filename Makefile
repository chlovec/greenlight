run:
	go run ./cmd/api
lint:
	golangci-lint run

migrate-create:
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "Usage: make migrate-create <file_name>"; \
		exit 1; \
	else \
		FILE_NAME=$(filter-out $@,$(MAKECMDGOALS)); \
		migrate create -seq -ext=.sql -dir=./migrations $$FILE_NAME; \
	fi

%:
	@:

migrate-up:
	@if [ -z "$(GREENLIGHT_DB_DSN)" ]; then \
		echo "Usage: GREENLIGHT_DB_DSN must be set or passed as a flag"; \
		exit 1; \
	fi
	migrate -path=./migrations -database="$(GREENLIGHT_DB_DSN)" up

migrate-down:
	@if [ -z "$(GREENLIGHT_DB_DSN)" ]; then \
		echo "Usage: GREENLIGHT_DB_DSN must be set or passed as a flag"; \
		exit 1; \
	fi
	migrate -path=./migrations -database="$(GREENLIGHT_DB_DSN)" down

migrate-force:
	@if [ -z "$(GREENLIGHT_DB_DSN)" ]; then \
		echo "Usage: GREENLIGHT_DB_DSN must be set or passed as a flag"; \
		exit 1; \
	fi
	migrate -path=./migrations -database="$(GREENLIGHT_DB_DSN)" force $(version)

