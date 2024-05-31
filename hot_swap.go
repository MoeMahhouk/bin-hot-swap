package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	hashBinaryFilePath = "./config/hash_binary.txt"
)

func main() {
	// Parse command-line arguments
	maxSwaps := flag.Int("max-swaps", 0, "maximum number of times binary swapping is allowed")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "This program swaps binaries based on changes in a hash.")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Channel to receive signals for graceful termination
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle termination signals
	go func() {
		<-sigCh
		fmt.Println("Received termination signal. Exiting...")
		cancel() // Cancel the context
	}()

	// Read initial hash and binary path from file
	initialHash, binaryPath, err := readHashAndBinaryPath(hashBinaryFilePath)
	if err != nil {
		fmt.Printf("Error reading initial hash and binary path from file: %v\n", err)
		return
	}

	// Execute binary initially
	fmt.Printf("Executing binary %s...\n", binaryPath)
	cmd := executeBinary(binaryPath)

	// Create a separate done channel to signal the end of the loop
	done := make(chan bool)

	// Periodically check for hash changes and update or re-execute binary
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	swaps := 0
	for {
		select {
		case <-ticker.C:
			// Read current hash and binary path from file
			currentHash, currentBinaryPath, err := readHashAndBinaryPath(hashBinaryFilePath)
			if err != nil {
				fmt.Printf("Error reading current hash and binary path from file: %v\n", err)
				continue
			}

			// Compare current hash with initial hash
			if !bytes.Equal([]byte(currentHash), []byte(initialHash)) {
				// Kill the previous process associated with the binary execution
				if cmd != nil {
					err := killProcessGroup(cmd)
					if err != nil {
						fmt.Printf("Error killing previous process group: %v\n", err)
					}
				}

				// Re-execute the program with the new binary path
				fmt.Printf("Hash has changed, re-executing the program with binary %s...\n", currentBinaryPath)
				cmd = executeBinary(currentBinaryPath)
				initialHash = string(currentHash)
				swaps++
				// Check if maximum swaps has been reached
				if *maxSwaps > 0 && swaps >= *maxSwaps {
					fmt.Println("Maximum swaps reached. Only the currently swapped binary will resume execution!")
					// Write the PID of the currently swapped binary to a file
					if err := writePidToFile("current_pid.txt", cmd.Process.Pid); err != nil {
						fmt.Printf("Error writing PID to file: %v\n", err)
					}
					done <- true // Signal the end of the hot swap loop
					break
				}
			}
		case <-ctx.Done():
			// Terminate child processes when the context is canceled
			if cmd != nil {
				err := killProcessGroup(cmd)
				if err != nil {
					fmt.Printf("Error killing process group: %v\n", err)
				}
			}
			// Exit the loop and terminate the program
			return
		case <-done:
			// Exit the loop and terminate the program
			return
		}
	}
}

func writePidToFile(filePath string, pid int) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))
	return err
}

func readHashAndBinaryPath(filePath string) (string, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}
	splitData := strings.SplitN(string(data), " ", 2)
	if len(splitData) != 2 {
		return "", "", fmt.Errorf("invalid data in file")
	}
	return splitData[0], strings.TrimSpace(splitData[1]), nil
}

func executeBinary(binaryPath string) *exec.Cmd {
	cmd := exec.Command("go", "run", binaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set the process group ID to the new process
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error executing binary %s: %v\n", binaryPath, err)
		return nil
	}
	return cmd
}

func killProcessGroup(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return err
	}
	return syscall.Kill(-pgid, syscall.SIGTERM)
}
