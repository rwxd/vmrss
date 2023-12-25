package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	flagMonitor  = flag.Bool("m", false, "Monitor process")
	flagChild    = flag.Bool("c", true, "Show child processes")
	flagInterval = flag.Int("i", 500, "Interval in milliseconds")
	flagTime     = flag.Int("t", 0, "Runtime in seconds")
	flagSwap     = flag.Bool("s", false, "Show swap memory")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		fmt.Println("Usage: vmrss [options] <pid>")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid PID")
		os.Exit(1)
	}

	if !processExists(pid) {
		fmt.Println("Process does not exist")
		os.Exit(1)
	}

	if *flagMonitor {
		if *flagTime > 0 {
			time.AfterFunc(time.Duration(*flagTime)*time.Second, func() {
				os.Exit(0)
			})
		}
		for {
			if processExists(pid) {
				printVmrss(pid, getVmrss(pid), *flagChild)
				time.Sleep(time.Duration(*flagInterval) * time.Millisecond)
			} else {
				fmt.Println("Process does not exist")
				break
			}
		}
	} else {
		printVmrss(pid, getVmrss(pid), *flagChild)
	}
}

type processOutput struct {
	Pid  int
	Name string
	// Space is used for indentation
	Space int
	// Mem is in MB
	Mem  float64
	Swap float64
}

// processExists checks if a process exists.
// We are not using os.FindProcess, because it will return a process even if it does not exist
func processExists(pid int) bool {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid))
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return len(strings.Split(string(out), "\n")) > 2
}

func printVmrss(mainPid int, processes []processOutput, children bool) {
	for _, process := range processes {
		if children || process.Pid == mainPid {
			if *flagSwap {
				fmt.Printf("%*s%s(%d): %.2f MB | swap: %.2f MB\n", process.Space, "", process.Name, process.Pid, process.Mem, process.Swap)
			} else {
				fmt.Printf("%*s%s(%d): %.2f MB\n", process.Space, "", process.Name, process.Pid, process.Mem)
			}
		}
	}

	total := getVmrssTotal(processes)
	if *flagSwap {
		fmt.Printf("Total: %.2f MB | swap: %.2f MB\n", total, getVmrssSwapTotal(processes))
	} else {
		fmt.Printf("Total: %.2f MB\n", total)
	}
}

func getVmrssTotal(processes []processOutput) float64 {
	var total float64
	for _, process := range processes {
		total += process.Mem
	}

	return total
}

func getVmrssSwapTotal(processes []processOutput) float64 {
	var total float64
	for _, process := range processes {
		total += process.Swap
	}

	return total
}

func getVmrss(mainPid int) []processOutput {
	outputs := []processOutput{}
	arr := []interface{}{mainPid, 0}

	for len(arr) > 0 {
		// remove last element
		space := arr[len(arr)-1].(int)
		arr = arr[:len(arr)-1]
		pid := arr[len(arr)-1].(int)
		arr = arr[:len(arr)-1]

		if !processExists(pid) {
			continue
		}

		mem, _ := getProcessVmrss(pid)
		swap, _ := getProcessVmSwap(pid)
		name, _ := getProcessName(pid)

		outputs = append(outputs, processOutput{
			Pid:   pid,
			Name:  name,
			Space: space,
			Mem:   mem,
			Swap:  swap,
		})

		children := getProcessChildren(pid)

		// add children to array
		for _, child := range children {
			arr = append(arr, child, space+2)
		}
	}
	return outputs
}

func getProcessInfo(pid int, key string) (float64, error) {
	status, err := getProcessStatus(pid)
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(status))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key) {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				mem, _ := strconv.ParseFloat(fields[1], 64)
				return mem / 1024, nil
			}
		}
	}

	return 0, fmt.Errorf("%s not found for PID %d", key, pid)
}

func getProcessStatus(pid int) (string, error) {
	filePath := fmt.Sprintf("/proc/%d/status", pid)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func getProcessVmrss(pid int) (float64, error) {
	return getProcessInfo(pid, "VmRSS:")
}

func getProcessVmSwap(pid int) (float64, error) {
	return getProcessInfo(pid, "VmSwap:")
}

func getProcessName(pid int) (string, error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func getProcessChildren(pid int) []int {
	cmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var children []int
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		child, _ := strconv.Atoi(line)
		children = append(children, child)
	}

	return children
}
