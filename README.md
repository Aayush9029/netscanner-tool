<p align="center">
  <img src="assets/icon.png" width="128" alt="netscanner">
  <h1 align="center">netscanner</h1>
  <p align="center">Find open ports on your local network without nmap</p>
</p>

<p align="center">
  <a href="https://github.com/Aayush9029/netscanner-tool/releases/latest"><img src="https://img.shields.io/github/v/release/Aayush9029/netscanner-tool" alt="Release"></a>
  <a href="https://github.com/Aayush9029/netscanner-tool/blob/main/LICENSE"><img src="https://img.shields.io/github/license/Aayush9029/netscanner-tool" alt="License"></a>
</p>

## Install

```bash
brew install aayush9029/tap/netscanner
```

Or tap first:

```bash
brew tap aayush9029/tap
brew install netscanner
```

## Usage

```bash
netscanner                                  # smart prompt for this network
netscanner scan 192.168.1.0/24             # scan a subnet
netscanner scan 192.168.1.1 -p 22,80,443   # scan selected ports
netscanner 192.168.1.10 --json             # script-friendly output
netscanner --timeout 500ms -c 800          # tune scan speed
```

`netscanner` detects the active interface, suggests the current subnet, gateway,
nearby hosts, and a manual target entry, then runs a native multi-threaded TCP
port scan in a live terminal dashboard.

## License

MIT
