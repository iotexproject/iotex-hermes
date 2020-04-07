# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOLINT=golint
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

lint:
	go list ./... | grep -v /vendor/ | xargs $(GOLINT)

test: fmt
	$(GOTEST) -short -race ./...