run:
	go run ./cmd/api
lint:
	golangci-lint run

migrate-create:
	@if [ -z "$(firstword $(MAKECMDGOALS))" ]; then \
		echo "Usage: make migrate-create <file_name>"; \
	else \
		migrate create -seq -ext=.sql -dir=./migrations file_name=$(firstword $(MAKECMDGOALS)); \
	fi;

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

