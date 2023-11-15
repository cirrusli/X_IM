# Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=X_IM

all: test build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
#
# The REAL run entrance,e.g.:make gateway
gateway:
	$(GOCMD) run main.go gateway
logic:
	$(GOCMD) run main.go logic
occult:
	$(GOCMD) run main.go occult
router:
	$(GOCMD) run main.go router
# You need to run commands above
#
deps:
	$(GOGET) github.com/spf13/cobra
	$(GOGET) github.com/spf13/viper