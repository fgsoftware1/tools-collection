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

var (
	installFlag bool
	outputFile   string
	proxychains4 string = "/etc/proxychains.conf"
)

func init() {
	flag.BoolVar(&installFlag, "i", false, "Install the program on the system")
	flag.BoolVar(&installFlag, "install", false, "Install the program on the system")
	flag.StringVar(&outputFile, "o", "working_proxies.txt", "Output file name")
	flag.StringVar(&outputFile, "output", "working_proxies.txt", "Output file name")
	flag.Parse()

	if installFlag {
		install()
		os.Exit(0)
	}
}

func main() {
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

	if len(targets) == 0 {
		fmt.Println("No targets found in the file.")
		return
	}

	fmt.Println("Starting proxy checking...")

	var workingProxies []Target

	for _, target := range targets {
		isOnline := scanPort(target)
		if isOnline {
			fmt.Printf("%s[+]%s %s:%s\n", Green, Reset, target.IP, target.Port)
		} else {
			fmt.Printf("%s[-]%s %s:%s\n", Red, Reset, target.IP, target.Port)
		}
	}

	fmt.Println("Proxy checking complete.")

	if err := writeProxychainsFile(workingProxies); err != nil {
		fmt.Printf("Error writing Proxychains4 file: %s\n", err)
		return
	}

	fmt.Println("Working proxies saved to file:", outputFile)
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

		parts := strings.FieldsFunc(line, func(r rune) bool {
			return r == ':' || r == ' '
		})

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

func writeProxychainsFile(proxies []Target) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, proxy := range proxies {
		line := fmt.Sprintf("socks5 %s %s\n", proxy.IP, proxy.Port)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
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

	if err := ioutil.WriteFile(installPath, execContent, 0755); err != nil {
		fmt.Printf("Failed to write the executable to the installation path: %s\n", err)
		return
	}

	fmt.Println("Installation complete.")
}