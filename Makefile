SHELL = /bin/bash

# Package and target names
APP_NAME = $(shell pwd | sed 's:.*/::')
TARGET = $(APP_NAME)

# Go paths and programs
BIN    = $(GOPATH)/bin
GOLINT = $(BIN)/golint
GOBUILD = go build -o $(TARGET)
GOTEST = go test -v -cover

build: clean tests
	@echo "MAKE: Building..."
	$(GOBUILD)
	@echo "MAKE: Done building!"

tests:
	$(GOLINT)
	$(GOTEST)

clean:
	@echo "MAKE: Cleaning..."
	-$(shell rm -f $(TARGET))

run: build
	@echo "MAKE: Running the binary..."
	./$(TARGET)
