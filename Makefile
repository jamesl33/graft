OUTPUT_DIR="build/bin/"

PACKAGE=
TESTS=

all: build

build:
	@mkdir -p $(OUTPUT_DIR)
	@go build -o $(OUTPUT_DIR)

coverage:
	@go test ./$(PACKAGE)/... -run=$(TESTS) -count=1 -covermode=atomic -coverprofile=coverage.out -failfast -shuffle=on -tags=test && go tool cover -html=coverage.out

generate:
	@go generate ./$(PACKAGE)/...

lint:
	@golangci-lint run

test:
	@go test ./$(PACKAGE)/... -run=$(TESTS) -count=1 -cover -failfast -shuffle=on -tags=test

clean:
	@rm -rf build
	@rm -f coverage.out

.PHONY: all build clean coverage generate lint test
