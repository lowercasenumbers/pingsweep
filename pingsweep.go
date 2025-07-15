package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// PingResult struct to hold the result of a single ping operation
type PingResult struct {
	IP      net.IP
	Message string
	IsUp    bool
}

// Global ping command parameters based on OS
var (
	pingParamCount   string
	pingParamTimeout string
	pingTimeoutValue string
)

func init() {
	// Initialize ping command parameters based on the operating system
	switch runtime.GOOS {
	case "windows":
		pingParamCount = "-n"
		pingParamTimeout = "-w"
		pingTimeoutValue = "1000" // milliseconds
	default: // Linux, macOS, etc.
		pingParamCount = "-c"
		pingParamTimeout = "-W"
		pingTimeoutValue = "1" // seconds
	}
}

// pingHost pings a single IP address and sends the result to a channel.
// `suppressPingHostStdout` controls if pingHost itself prints verbose commands or errors.
func pingHost(ip net.IP, verbose bool, suppressPingHostStdout bool, results chan<- PingResult) {
	defer wg.Done() // Decrement the WaitGroup counter when the goroutine finishes

	cmd := exec.Command("ping", pingParamCount, "1", pingParamTimeout, pingTimeoutValue, ip.String())

	if verbose && !suppressPingHostStdout {
		fmt.Printf("Pinging %s with command: %s\n", ip.String(), strings.Join(cmd.Args, " "))
	}

	err := cmd.Run()

	if err == nil {
		results <- PingResult{IP: ip, Message: fmt.Sprintf("%s is UP", ip.String()), IsUp: true}
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			if verbose && !suppressPingHostStdout {
				results <- PingResult{IP: ip, Message: fmt.Sprintf("%s is DOWN or unreachable (Return Code: %d)", ip.String(), exitError.ExitCode()), IsUp: false}
			} else {
				results <- PingResult{IP: ip, Message: "", IsUp: false}
			}
		} else {
			if !suppressPingHostStdout {
				results <- PingResult{IP: ip, Message: fmt.Sprintf("An unexpected error occurred while pinging %s: %v", ip.String(), err), IsUp: false}
			} else {
				results <- PingResult{IP: ip, Message: "", IsUp: false}
			}
		}
	}
}

var wg sync.WaitGroup // WaitGroup to wait for all goroutines to complete

func main() {
	// 1. Command-line argument parsing
	var (
		verbose     bool
		numThreads  int
		onlyIPs     bool   // -i flag: print only IPs to stdout, suppress other stdout info
		outputFile  string // -o flag: save IPs to file
		networkCIDR string
	)

	// Combined flag definitions for cleaner help output
	flag.BoolVar(&verbose, "v", false, "(-v, --verbose) Enable verbose output. Shows ping commands and return codes. This is suppressed if -i is used.")
	flag.IntVar(&numThreads, "t", 10, "(-t, --threads) The maximum number of concurrent ping operations to run (number of threads). (Default: 10)")
	flag.BoolVar(&onlyIPs, "i", false, "(-i, --only-ips) Output only the IP addresses of UP hosts to standard output, one per line. Suppresses banners and 'DOWN' messages to stdout.")
	flag.StringVar(&outputFile, "o", "", "(-o, --output) Save the IP addresses of UP hosts to the specified FILE, one per line. When used, results are still printed to stdout.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <network_cidr>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -v -t 20 192.168.1.0/24\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -o up_hosts.txt 192.168.1.0/24 (Outputs to file AND stdout)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -i 192.168.1.0/24 (Outputs ONLY IPs to stdout)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nThis script performs a concurrent IP ping sweep on a specified network range.\n")
		fmt.Fprintf(os.Stderr, "It reports hosts that are found to be reachable ('UP').\n")
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  <network_cidr>  The network address in CIDR notation (e.g., \"192.168.1.0/24\").\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults() // This will now print the combined help strings
	}

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: Missing network_cidr argument.")
		flag.Usage()
		os.Exit(1)
	}
	networkCIDR = flag.Arg(0)

	// suppressNonIPInfoForStdout: True if -i is used. Controls banners, verbose messages, and "DOWN" messages to stdout.
	suppressNonIPInfoForStdout := onlyIPs

	var fileWriter io.Writer // Will be nil if no output file is specified
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Error: Could not open output file '%s': %v\n", outputFile, err)
		}
		defer f.Close() // Ensure the file is closed when main exits
		fileWriter = f
	}

	// Print initial banners to stdout, unless suppressNonIPInfoForStdout is true
	if !suppressNonIPInfoForStdout {
		fmt.Printf("Starting concurrent ping sweep for network: %s\n", networkCIDR)
		fmt.Printf("Using up to %d concurrent threads.\n", numThreads)
		fmt.Println("-----------------------------------")
	}

	// Parse the CIDR network
	_, ipnet, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		fmt.Printf("Error: Invalid network address or CIDR notation: %v\n", err)
		os.Exit(1)
	}

	ip4 := ipnet.IP.To4()
	if ip4 == nil {
		fmt.Printf("Error: Only IPv4 networks are supported for iteration. Provided: %s\n", networkCIDR)
		os.Exit(1)
	}

	networkUint32 := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])

	ones, bits := ipnet.Mask.Size()
	if bits != 32 {
		fmt.Printf("Error: Only IPv4 /XX subnets are supported for iteration. Provided: %s\n", networkCIDR)
		os.Exit(1)
	}

	numAddressesInSubnet := uint32(1 << (uint(bits - ones)))

	firstUsableHostUint32 := networkUint32 + 1
	lastUsableHostUint32 := networkUint32 + numAddressesInSubnet - 2

	results := make(chan PingResult, numThreads)
	var foundUpHosts bool

	if numAddressesInSubnet > 2 { // Only iterate if there are usable hosts
		for i := firstUsableHostUint32; i <= lastUsableHostUint32; i++ {
			currentIP := make(net.IP, 4)
			currentIP[0] = byte(i >> 24)
			currentIP[1] = byte(i >> 16)
			currentIP[2] = byte(i >> 8)
			currentIP[3] = byte(i)

			if ipnet.Contains(currentIP) { // Safety check
				wg.Add(1)
				// Pass `suppressNonIPInfoForStdout` to `pingHost`
				// This ensures pingHost itself doesn't print verbose commands or errors if -i is active.
				go pingHost(currentIP, verbose, suppressNonIPInfoForStdout, results)
			}
		}
	} else {
		if !suppressNonIPInfoForStdout {
			fmt.Println("No usable hosts in this subnet (e.g., /31 or /32).")
		}
	}

	// Goroutine to close the results channel once all pings are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results from the channel
	for res := range results {
		if res.IsUp {
			foundUpHosts = true
			// Always write to file if file output is enabled
			if fileWriter != nil {
				fmt.Fprintf(fileWriter, "%s\n", res.IP.String())
			}

			// Print to stdout based on `onlyIPs` flag
			if onlyIPs {
				// If -i is used, print only IP to stdout
				fmt.Println(res.IP.String())
			} else {
				// If -i is NOT used, print full message to stdout
				fmt.Println(res.Message)
			}
		} else { // Host is DOWN
			// Print DOWN message to stdout only if verbose AND not suppressing non-IP info
			if res.Message != "" && verbose && !suppressNonIPInfoForStdout {
				fmt.Println(res.Message)
			}
		}
	}

	// Final messages to stdout
	if !foundUpHosts && !suppressNonIPInfoForStdout {
		fmt.Println("No UP hosts found in the specified range.")
		if !verbose { // Suggest -v only if it wasn't used
			fmt.Println("Run with -v or --verbose for more details on unreachable hosts.")
		}
	}

	if !suppressNonIPInfoForStdout {
		fmt.Println("-----------------------------------")
		fmt.Println("Concurrent ping sweep finished.")
	}
}

