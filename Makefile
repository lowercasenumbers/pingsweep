# Define variables
# The name of your Go source file
SRC_FILE = pingsweep.go
# The desired name for your compiled executable
BIN_NAME = pingsweep
# The directory where the executable will be built within the project
BUILD_DIR = bin
# The full path to the built executable
BUILD_PATH = $(BUILD_DIR)/$(BIN_NAME)
# The directory where the executable will be installed system-wide
INSTALL_DIR = $(HOME)/.local/bin

# .PHONY targets are not actual files; they are commands to be executed.
.PHONY: all build install clean help

# Default target: builds the executable
all: build

# Build target: compiles the Go program into the BUILD_DIR
build:
	@echo "Building $(BIN_NAME)..."
	# Create the build directory if it doesn't already exist
	mkdir -p $(BUILD_DIR)
	# go mod tidy ensures your go.mod and go.sum are up-to-date and dependencies are downloaded.
	# It's good practice even for single-file projects.
	go mod tidy
	# Compile the Go source file into an executable named $(BIN_NAME) inside $(BUILD_DIR)
	go build -o $(BUILD_PATH) $(SRC_FILE)
	@echo "Build complete: ./"$(BUILD_PATH)

# Install target: builds the executable and copies it to the install directory
install: build
	@echo "Installing $(BIN_NAME) from $(BUILD_PATH) to $(INSTALL_DIR)..."
	# Create the installation directory if it doesn't already exist
	mkdir -p $(INSTALL_DIR)
	# Copy the compiled executable from the build directory to the installation directory
	cp $(BUILD_PATH) $(INSTALL_DIR)/$(BIN_NAME)
	@echo "Installation complete. You can now run '$(BIN_NAME)' from anywhere."
	@echo "Make sure $(INSTALL_DIR) is in your PATH environment variable."

# Clean target: removes the compiled executable from the build directory
clean:
	@echo "Cleaning up..."
	# Remove the build directory and its contents
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Help target: displays usage information for the Makefile
help:
	@echo "Makefile for $(BIN_NAME)"
	@echo ""
	@echo "Usage:"
	@echo "  make build    - Builds the executable into the '$(BUILD_DIR)/' directory."
	@echo "  make install  - Builds the executable and copies it to $(INSTALL_DIR)."
	@echo "  make clean    - Removes the built executable and its directory."
	@echo "  make help     - Displays this help message."
