
.PHONY: build
build: clean
	@echo "Building binary"
	go build -o bin/mars cmd/mars/main.go

.PHONY: build
clean:
	@echo "Cleaning bin directory"
	@rm bin/* || true
