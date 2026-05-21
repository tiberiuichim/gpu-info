BINARY=gpu-info
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: build clean

build:
	go build -o $(BINARY) .
	install -d $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/

clean:
	rm -f $(BINARY)
