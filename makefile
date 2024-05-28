BINARY_NAME=rewit
INSTALL_DIR=/usr/local/bin

# Default target
all: build

# Build the Go binary
build:
	go build -o $(BINARY_NAME)

# Install the binary to INSTALL_DIR
install: build
	sudo mv $(BINARY_NAME) $(INSTALL_DIR)

# Clean up the build artifacts
clean:
	rm -f $(BINARY_NAME)

# Help message
help:
	@echo "Rewit"
	@echo ""
	@echo "Usage:"
	@echo "  make          Build the project"
	@echo "  make build    Build the project"
	@echo "  make install  Install the binary system-wide"
	@echo "  make clean    Clean up build artifacts"
	@echo "  make help     Show this help message"

