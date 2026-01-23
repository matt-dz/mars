
.PHONY: build
build: clean
	@echo "Building binary..."
	go build -o bin/mars cmd/mars/main.go

.PHONY: build
clean:
	@echo "Cleaning bin directory..."
	@rm bin/* || true

.PHONY: fmt
fmt: sql-fmt
	@echo "Formatting..."
	golangci-lint run

.PHONY: lint
lint:
	@echo "Linting..."
	golangci-lint run

.PHONY: sql-fmt
sql-fmt:
	@echo "Formatting sql..."
	pg_format -i internal/database/sql/schema.sql internal/database/sql/query.sql

.PHONY: sqlc
sqlc: sql-fmt
	@echo "Generating sqlc files..."
	sqlc generate

.PHONY: tidy
tidy:
	@echo "Tidying go..."
	go mod tidy
