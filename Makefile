BINARY=gpu-info
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: all build install clean

all: build install

build:
	go build -o $(BINARY) .

install:
	install -d $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/

clean:
	rm -f $(BINARY)
