.PHONY: build run test clean help

APP_NAME=gobi
MAIN_PATH=cmd/gobi/main.go

help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the build and run script"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Remove binary and clean go cache"
	@echo "  make help          - Show this help message"

build:
	go build -o $(APP_NAME) $(MAIN_PATH)

run:
	./build_and_run.sh

test:
	go test ./...

clean:
	rm -f $(APP_NAME)
	go clean
