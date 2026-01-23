
.PHONY: build
build: clean
	@echo "Building binary"
	go build -o bin/mars cmd/mars/main.go

.PHONY: build
clean:
	@echo "Cleaning bin directory"
	@rm bin/* || true

.PHONY: fmt
fmt:
	@echo "Formatting sql"
	pg_format -i internal/database/sql/schema.sql
