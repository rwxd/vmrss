# vmrss

A simple tool to show the memory usage of a process and its children.

## Install

Get the latest binary from the [releases page](https://github.com/rwxd/vmrss/releases).

Or install with go:

```bash
go get github.com/rwxd/vmrss
```

### NixOS / Nix

Add to your `flake.nix`:

```nix
{
  inputs.vmrss.url = "github:rwxd/vmrss";
  # Or use a specific version:
  # inputs.vmrss.url = "github:rwxd/vmrss/v1.0.0";
}
```

Then use in your configuration:

```nix
environment.systemPackages = [ inputs.vmrss.packages.${system}.vmrss ];
```

Or run directly:

```bash
nix run github:rwxd/vmrss -- -m <pid>
# Or use a specific version:
# nix run github:rwxd/vmrss/v1.0.0 -- -m <pid>
```

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
