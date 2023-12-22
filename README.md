# vmrss

A simple tool to show the memory usage of a process and its children.

## Usage

### Continuously monitor a process

```bash
vmrss -m 3840
.kitty-wrapped(3840): 151.23 MB
  zsh(3851): 12.50 MB
    tmux: client(3992): 5.20 MB
    zsh(3938): 9.28 MB
Total: 178.20 MB
```

### Set custom interval in milliseconds

```bash
vmrss -m -i 2000 3840
```

### Quit after a certain amount of time in seconds

```bash
vmrss -m -t 10 26847
```

### Do not show children

```bash
vmrss -m -c=false 26847
```

### Show swapped memory

```bash
vmrss -m -s 3840
.kitty-wrapped(3840): 151.36 MB | swap: 0.00 MB
  zsh(3851): 12.50 MB | swap: 0.00 MB
    tmux: client(3992): 5.20 MB | swap: 0.00 MB
    zsh(3938): 9.28 MB | swap: 0.00 MB
Total: 178.34 MB | swap: 0.00 MB
```
