OUTPUT_DIR="build/bin/"

COMMANDS=$(shell ls ./cmd)
BINARIES?=$(COMMANDS)

PACKAGE=
TESTS=

all: build

build:
	@mkdir -p $(OUTPUT_DIR)
	@for binary in $(BINARIES); do go build -ldflags $(LD_FLAGS) -o $(OUTPUT_DIR) ./cmd/$$binary; done

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
