NAME      = mallard
BUILD_DIR = $(CURDIR)/build

.PHONY: all clean
all: clean server

clean:
	rm -rf $(BUILD_DIR)

server: server-linux-amd64 server-darwin-amd64 server-darwin-arm64

server-linux-amd64 server-darwin-amd64 server-darwin-arm64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$$(echo $@ | sed 's/^server-//;s/-.*//') GOARCH=$$(echo $@ | sed 's/^server-//;s/.*-//') \
		go build -o $(BUILD_DIR)/$(NAME)-$@ ./cmd/server
