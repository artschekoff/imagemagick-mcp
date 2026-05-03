VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BINARY   := imagemagick-mcp
BIN_DIR  := bin
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

.PHONY: build build-all test clean release

build:
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) .

build-all:
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY)-windows-amd64.exe .

test:
	go test ./... -v

clean:
	rm -rf $(BIN_DIR)

install: build
	sudo cp $(BIN_DIR)/$(BINARY) /usr/local/bin/$(BINARY)

release: build-all
	gh release create $(VERSION) $(BIN_DIR)/* --generate-notes --title "Release $(VERSION)"
