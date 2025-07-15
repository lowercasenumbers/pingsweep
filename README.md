# pingsweep 

`pingsweep` is a fast, concurrent IP ping sweep tool written in Go. It allows you to quickly identify active hosts within a specified CIDR network range, with options for verbose output, controlling concurrency, and directing results to standard output or a file.

## Features

* **Concurrent Scanning:** Utilizes Go's goroutines and channels for high-speed parallel ping operations.
* **CIDR Support:** Scans any valid IPv4 CIDR range (e.g., `192.168.1.0/24`, `10.0.0.0/16`).
* **Customizable Concurrency:** Control the number of simultaneous ping requests using the `--threads` option.
* **Flexible Output:**
    * By default, reports "UP" hosts to standard output.
    * **Verbose Mode (`-v` / `--verbose`):** Shows detailed information, including ping commands executed and return codes for unreachable hosts.
    * **Only IPs Mode (`-i` / `--only-ips`):** Outputs only the IP addresses of reachable hosts, one per line, ideal for piping to other tools like Nmap.
    * **Output to File (`-o` / `--output`):** Saves the IP addresses of reachable hosts to a specified file while still printing "UP" messages to standard output.

## Installation

To build and install `pingsweep`, you'll need the Go programming language installed on your system.

1.  **Save the script:** Make sure your Go code is saved into a file named `pingsweep.go`.
2.  **Create a `Makefile`:** In the **same directory** as `pingsweep.go`, create a file named `Makefile` with the following content:

    ```makefile
    # Define variables
    SRC_FILE = pingsweep.go
    BIN_NAME = pingsweep
    INSTALL_DIR = $(HOME)/.local/bin

    .PHONY: all build install clean help

    all: build

    build:
    	@echo "Building $(BIN_NAME)..."
    	go mod tidy
    	go build -o $(BIN_NAME) $(SRC_FILE)
    	@echo "Build complete: ./"$(BIN_NAME)

    install: build
    	@echo "Installing $(BIN_NAME) to $(INSTALL_DIR)..."
    	mkdir -p $(INSTALL_DIR)
    	cp $(BIN_NAME) $(INSTALL_DIR)/$(BIN_NAME)
    	@echo "Installation complete. You can now run '$(BIN_NAME)' from anywhere."
    	@echo "Make sure $(INSTALL_DIR) is in your PATH environment variable."

    clean:
    	@echo "Cleaning up..."
    	@rm -f $(BIN_NAME)
    	@echo "Clean complete."

    help:
    	@echo "Makefile for $(BIN_NAME)"
    	@echo ""
    	@echo "Usage:"
    	@echo "  make build    - Builds the executable in the current directory."
    	@echo "  make install  - Builds the executable and copies it to $(INSTALL_DIR)."
    	@echo "  make clean    - Removes the built executable from the current directory."
    	@echo "  make help     - Displays this help message."
    ```

3.  **Initialize Go Module:** Open your terminal, navigate to the directory where you saved the files, and initialize a Go module. You can use any name for your module, e.g., `pingsweep_project`.

    ```bash
    go mod init pingsweep_project
    ```

4.  **Build and Install:** Use the `Makefile` to build and install the executable to `~/.local/bin/`.

    ```bash
    make install
    ```

    This command will:
    * Ensure Go module dependencies are clean (`go mod tidy`).
    * Compile `pingsweep.go` into an executable named `pingsweep`.
    * Create the `~/.local/bin/` directory if it doesn't exist.
    * Copy the `pingsweep` executable to `~/.local/bin/`.

    **Note:** For `pingsweep` to be runnable from any directory, ensure `~/.local/bin/` is in your system's `PATH` environment variable. Most modern Linux/macOS systems configure this by default. If not, add `export PATH="$HOME/.local/bin:$PATH"` to your shell's configuration file (e.g., `~/.bashrc`, `~/.zshrc`) and then `source` it or open a new terminal.

## Usage

All options must be placed **before** the network CIDR argument.

```bash
pingsweep [options] <network_cidr>
