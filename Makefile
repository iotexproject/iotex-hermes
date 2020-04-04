# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BUILD_TARGET_SERVER=hermes

install:
	$(GOCMD) install -v

build: clean
	$(GOBUILD) -o ./bin/$(BUILD_TARGET_SERVER)

run: build
	./bin/$(BUILD_TARGET_SERVER)

clean:
	@echo "Cleaning..."
	rm -rf ./bin/$(BUILD_TARGET_SERVER)

fmt:
	$(GOCMD) fmt ./...

test: fmt
	$(GOTEST) -short -race ./...