package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

const (
	Reset = "\033[0m"
	Green = "\033[32m"
	Red   = "\033[31m"
)

type Target struct {
	IP   string
	Port string
}

var installFlag bool

func init() {
	flag.BoolVar(&installFlag, "install", false, "Install the program on the system")
	flag.Parse()
}

func main() {
	if installFlag {
		install()
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run proxychecker.go [proxyfile]")
		return
	}

	targetFile := os.Args[1]

	targets, err := readTargetsFromFile(targetFile)
	if err != nil {
		fmt.Printf("Error reading targets from file: %s\n", err)
		return
	}

	fmt.Println("Starting proxy checking...")

	for _, target := range targets {
		isOnline := scanPort(target)
		if isOnline {
			fmt.Printf("%s[+]%s %s:%s\n", Green, Reset, target.IP, target.Port)
		} else {
			fmt.Printf("%s[-]%s %s:%s\n", Red, Reset, target.IP, target.Port)
		}
	}

	fmt.Println("Proxy checking complete.")
}

func readTargetsFromFile(filePath string) ([]Target, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []Target
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}

		target := Target{
			IP:   parts[0],
			Port: parts[1],
		}

		targets = append(targets, target)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

func scanPort(target Target) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", target.IP, target.Port), 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

func install() {
	fmt.Println("Installing the program...")

	installDir := "/usr/local/bin"
	executableName := "proxychecker"

	currentExecPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get the current executable path: %s\n", err)
		return
	}

	execContent, err := ioutil.ReadFile(currentExecPath)
	if err != nil {
		fmt.Printf("Failed to read the current executable: %s\n", err)
		return
	}

	installPath := fmt.Sprintf("%s/%s", installDir, executableName)

	if _, err := os.Stat(installPath); err == nil {
		fmt.Printf("Program already installed at: %s\n", installPath)
		return
	}

	if err := ioutil.WriteFile(installPath, execContent, 0755); err != nil {
		fmt.Printf("Failed to write the executable to the installation path: %s\n", err)
		return
	}

	fmt.Println("Installation complete.")
}
