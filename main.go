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
	flagInterval = flag.String("i", "1s", "Interval (e.g., 500ms, 2s, 1m)")
	flagTime     = flag.String("t", "", "Quit after duration (e.g., 5s, 1m)")
	flagSwap     = flag.Bool("swap", false, "Show swap memory")
	flagCPU      = flag.Bool("cpu", false, "Show CPU usage")
	flagPeak     = flag.Bool("peak", false, "Show peak memory usage")
	flagIO       = flag.Bool("io", false, "Show disk I/O rates")
)

func main() {
	flag.Parse()
	args := flag.Args()

	var pids []int

	if len(args) == 0 {
		fmt.Println("Usage: vmrss [options] <pid|name> [<pid|name>...]")
		os.Exit(1)
	}

	for _, arg := range args {
		pid, err := strconv.Atoi(arg)
		if err != nil {
			foundPids, err := findProcessByName(arg)
			if err != nil || len(foundPids) == 0 {
				fmt.Printf("No process found with name: %s\n", arg)
				os.Exit(1)
			}
			pids = append(pids, filterRootProcesses(foundPids)...)
		} else {
			pids = append(pids, pid)
		}
	}

	peakMemory := make(map[int]float64)
	peakTotal := make(map[int]float64)
	lastIO := make(map[int][2]float64) // [readBytes, writeBytes]
	lastIOTime := time.Now()

	interval, err := time.ParseDuration(*flagInterval)
	if err != nil {
		fmt.Printf("Invalid interval format: %s\n", *flagInterval)
		os.Exit(1)
	}

	if *flagMonitor {
		if *flagTime != "" {
			timeout, err := time.ParseDuration(*flagTime)
			if err != nil {
				fmt.Printf("Invalid timeout format: %s\n", *flagTime)
				os.Exit(1)
			}
			time.AfterFunc(timeout, func() {
				os.Exit(0)
			})
		}

		for {
			now := time.Now()
			elapsed := now.Sub(lastIOTime).Seconds()
			displayProcesses(pids, peakMemory, peakTotal, lastIO, elapsed)
			lastIOTime = now
			fmt.Println()
			time.Sleep(interval)
		}
	} else {
		displayProcesses(pids, peakMemory, peakTotal, lastIO, 0)
	}
}

func displayProcesses(pids []int, peakMemory, peakTotal map[int]float64, lastIO map[int][2]float64, elapsed float64) {
	for i, pid := range pids {
		if i > 0 {
			fmt.Println()
		}
		processes := getVmrss(pid, peakMemory, lastIO, elapsed)
		if len(processes) == 0 {
			continue
		}
		currentTotal := getVmrssTotal(processes)
		if currentTotal > peakTotal[pid] {
			peakTotal[pid] = currentTotal
		}
		printVmrss(pid, processes, *flagChild, peakTotal[pid])
	}
}

type processOutput struct {
	Pid       int
	Name      string
	Space     int     // Indentation level
	Mem       float64 // Memory in MB
	Swap      float64 // Swap in MB
	CPU       float64 // CPU usage percentage
	PeakMem   float64 // Peak memory in MB
	ReadRate  float64 // Disk read rate in KB/s
	WriteRate float64 // Disk write rate in KB/s
}

// os.FindProcess returns a process even if it doesn't exist, so we check via /proc
func getProcessInfo(pid int, key string) (float64, error) {
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
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

func printVmrss(mainPid int, processes []processOutput, children bool, peakTotal float64) {
	for _, process := range processes {
		if children || process.Pid == mainPid {
			output := fmt.Sprintf("%*s%s(%d): %.2f MB", process.Space, "", process.Name, process.Pid, process.Mem)

			if *flagPeak && process.PeakMem > 0 {
				output += fmt.Sprintf(" | peak: %.2f MB", process.PeakMem)
			}

			if *flagCPU {
				output += fmt.Sprintf(" | cpu: %.1f%%", process.CPU)
			}

			if *flagIO {
				if *flagMonitor {
					output += fmt.Sprintf(" | io: r %.1f KB/s w %.1f KB/s", process.ReadRate, process.WriteRate)
				} else {
					output += fmt.Sprintf(" | io: r %.2f MB w %.2f MB", process.ReadRate/1024, process.WriteRate/1024)
				}
			}

			if *flagSwap {
				output += fmt.Sprintf(" | swap: %.2f MB", process.Swap)
			}

			fmt.Println(output)
		}
	}

	total := getVmrssTotal(processes)
	output := fmt.Sprintf("total: %.2f MB", total)

	if *flagPeak && peakTotal > 0 {
		output += fmt.Sprintf(" | peak: %.2f MB", peakTotal)
	}

	if *flagCPU {
		output += fmt.Sprintf(" | cpu: %.1f%%", getVmrssCPUTotal(processes))
	}

	if *flagIO {
		totalRead, totalWrite := getVmrssIOTotal(processes)
		if *flagMonitor {
			output += fmt.Sprintf(" | io: r %.1f KB/s w %.1f KB/s", totalRead, totalWrite)
		} else {
			output += fmt.Sprintf(" | io: r %.2f MB w %.2f MB", totalRead/1024, totalWrite/1024)
		}
	}

	if *flagSwap {
		output += fmt.Sprintf(" | swap: %.2f MB", getVmrssSwapTotal(processes))
	}

	fmt.Println(output)
}

func getVmrssTotal(processes []processOutput) float64 {
	var total float64
	for _, p := range processes {
		total += p.Mem
	}
	return total
}

func getVmrssSwapTotal(processes []processOutput) float64 {
	var total float64
	for _, p := range processes {
		total += p.Swap
	}
	return total
}

func getVmrssCPUTotal(processes []processOutput) float64 {
	var total float64
	for _, p := range processes {
		total += p.CPU
	}
	return total
}

func getVmrssIOTotal(processes []processOutput) (float64, float64) {
	var totalRead, totalWrite float64
	for _, p := range processes {
		totalRead += p.ReadRate
		totalWrite += p.WriteRate
	}
	return totalRead, totalWrite
}

func getVmrss(mainPid int, peakMemory map[int]float64, lastIO map[int][2]float64, elapsed float64) []processOutput {
	var outputs []processOutput
	arr := []any{mainPid, 0}

	for len(arr) > 0 {
		space := arr[len(arr)-1].(int)
		arr = arr[:len(arr)-1]
		pid := arr[len(arr)-1].(int)
		arr = arr[:len(arr)-1]

		mem, err := getProcessInfo(pid, "VmRSS:")
		if err != nil {
			continue
		}

		swap, _ := getProcessInfo(pid, "VmSwap:")
		name, cpu := getProcessNameAndCPU(pid)

		readRate, writeRate := 0.0, 0.0
		if *flagIO {
			readRate, writeRate = getProcessIORate(pid, lastIO, elapsed)
		}

		if mem > peakMemory[pid] {
			peakMemory[pid] = mem
		}

		outputs = append(outputs, processOutput{
			Pid:       pid,
			Name:      name,
			Space:     space,
			Mem:       mem,
			Swap:      swap,
			CPU:       cpu,
			PeakMem:   peakMemory[pid],
			ReadRate:  readRate,
			WriteRate: writeRate,
		})

		for _, child := range getProcessChildren(pid) {
			arr = append(arr, child, space+2)
		}
	}
	return outputs
}

func getProcessNameAndCPU(pid int) (string, float64) {
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return "", 0
	}

	stat := string(content)
	fields := strings.Fields(stat)
	if len(fields) < 22 {
		return "", 0
	}

	name := ""
	if start := strings.Index(stat, "("); start >= 0 {
		if end := strings.LastIndex(stat, ")"); end > start {
			name = stat[start+1 : end]
		}
	}

	cpu := 0.0
	if *flagCPU {
		utime, _ := strconv.ParseFloat(fields[13], 64)
		stime, _ := strconv.ParseFloat(fields[14], 64)
		if uptime, err := getSystemUptime(); err == nil {
			starttime, _ := strconv.ParseFloat(fields[21], 64)
			hertz := 100.0
			if seconds := uptime - (starttime / hertz); seconds > 0 {
				cpu = ((utime + stime) / hertz / seconds) * 100
			}
		}
	}

	return name, cpu
}

func getProcessChildren(pid int) []int {
	cmd := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var children []int
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		if child, err := strconv.Atoi(line); err == nil {
			children = append(children, child)
		}
	}
	return children
}

func findProcessByName(name string) ([]int, error) {
	cmd := exec.Command("pgrep", "-i", name)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var pids []int
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			if pid, err := strconv.Atoi(line); err == nil {
				pids = append(pids, pid)
			}
		}
	}
	return pids, nil
}

func filterRootProcesses(pids []int) []int {
	childSet := make(map[int]bool)
	for _, pid := range pids {
		for _, child := range getProcessChildren(pid) {
			childSet[child] = true
		}
	}

	var roots []int
	for _, pid := range pids {
		if !childSet[pid] {
			roots = append(roots, pid)
		}
	}
	return roots
}

func getSystemUptime() (float64, error) {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(content))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}
	return strconv.ParseFloat(fields[0], 64)
}

func getProcessIORate(pid int, lastIO map[int][2]float64, elapsed float64) (float64, float64) {
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/io", pid))
	if err != nil {
		return 0, 0
	}

	var readBytes, writeBytes float64
	for line := range strings.SplitSeq(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			val, _ := strconv.ParseFloat(fields[1], 64)
			if fields[0] == "read_bytes:" {
				readBytes = val
			} else if fields[0] == "write_bytes:" {
				writeBytes = val
			}
		}
	}

	if elapsed == 0 {
		lastIO[pid] = [2]float64{readBytes, writeBytes}
		return readBytes / 1024, writeBytes / 1024
	}

	last, exists := lastIO[pid]
	lastIO[pid] = [2]float64{readBytes, writeBytes}

	if !exists {
		return 0, 0
	}

	readRate := (readBytes - last[0]) / elapsed / 1024
	writeRate := (writeBytes - last[1]) / elapsed / 1024

	if readRate < 0 {
		readRate = 0
	}
	if writeRate < 0 {
		writeRate = 0
	}

	return readRate, writeRate
}
