# vmrss

Monitor memory usage of processes and their children.

## Requirements

- Linux with `/proc` filesystem
- `pgrep` command (usually from `procps` or `procps-ng` package)

## Install

```bash
go install github.com/rwxd/vmrss@latest
```

Or download from [releases](https://github.com/rwxd/vmrss/releases).

## Usage

```bash
# By PID
vmrss 1234

# By name
vmrss firefox

# Multiple processes
vmrss firefox discord

# Monitor continuously
vmrss -m firefox

# With CPU and peak memory
vmrss -m -cpu -peak firefox
```

## Options

```
-m          Monitor continuously (default: single snapshot)
-i          Interval (default: 1s, examples: 500ms, 2s, 1m)
-t          Quit after duration (examples: 5s, 1m, 30s)
-c          Show child processes (default: true)
-cpu        Show CPU usage
-peak       Show peak memory usage
-swap       Show swap memory
-io         Show disk I/O rates (KB/s)
```

## Examples

```bash
# Monitor with 500ms interval
vmrss -m -i 500ms firefox

# Show only parent process
vmrss -c=false firefox

# Monitor for 10 seconds with CPU and peak
vmrss -m -t 10s -cpu -peak discord

# Show disk I/O rates
vmrss -m -io firefox
```
