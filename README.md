# vmrss

A simple tool to show the memory usage of a process and its children.

## Usage

### Continously monitor a process

```bash
vmrss -m 26847
python(26847): 43.85 MB
  python(26849): 48.44 MB
Total: 92.29 MB
```

### Set custom interval in milliseconds

```bash
vmrss -m -i 2000 3840
.kitty-wrapped(3840): 151.23 MB
  zsh(3851): 12.50 MB
    tmux: client(3992): 5.20 MB
    zsh(3938): 9.28 MB
Total: 178.20 MB
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
.kitty-wrapped(3840): 151.36 MB (0.00 MB swap)
  zsh(3851): 12.50 MB (0.00 MB swap)
    tmux: client(3992): 5.20 MB (0.00 MB swap)
    zsh(3938): 9.28 MB (0.00 MB swap)
Total: 178.33 MB (0.00 MB swap)
```
