# Network Scanner

Network Scanner is an async CLI utility for macOS that inspects your local network, discovers devices through the ARP table, and uses `nmap` to report their open ports in a Rich-powered UI.

<div align="center">

![Python](https://img.shields.io/badge/python-3.9%2B-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Version](https://img.shields.io/badge/version-0.2.0-orange.svg)

</div>

## Features

- Discovers hosts via the system ARP table (no noisy network sweep required)
- Async port discovery backed by `nmap` for accurate results
- MAC-to-vendor lookup using the macvendors.com API
- Rich terminal experience with progress indicators, banners, and summary tables
- Configurable port lists and concurrency settings

![Demo](./resources/demo.gif)

## Prerequisites

| Requirement | Notes |
| --- | --- |
| macOS 13+ | Uses the built-in `arp` command that ships with macOS |
| Python 3.9+ | Any modern CPython release works; 3.12+ is recommended |
| `nmap` | `brew install nmap` |
| Internet access | Needed for vendor lookups (can be skipped with `--no-check`) |

## Install & Run

1. **Clone the repo**
   ```bash
   git clone https://github.com/Aayush9029/netscanner-tool.git
   cd netscanner-tool
   ```
2. **Pick one of the workflows below.** Each keeps the install isolated so reproducing issues is trivial (e.g., create a `sandbox` folder and follow the same steps there when testing the README).

### Option A: pip + `venv` (recommended)

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install --upgrade pip
pip install -e .          # Use `pip install .` if you do not need editable installs
```

### Option B: uv

```bash
uv venv                    # Creates .venv using the Python uv ships with
source .venv/bin/activate
uv pip install -e .        # Or use `uv pip install --system -e .` without a venv
```

> **Tip:** `uv pip install -e .` requires an active virtual environment. If you skip `uv venv`, pass `--system` so uv installs into the global interpreter instead.

### Run the scanner

```bash
netscanner --help
netscanner                 # Basic scan (with dependency checks)
netscanner --no-check      # Skip the nmap/arp checks if you know they exist
```

## Usage Examples

```bash
netscanner --ports 22,80,443    # Scan only these ports
netscanner --max-concurrent 10  # Raise concurrency for faster runs
netscanner --no-check           # Skip dependency verification
```

By default the tool asks `nmap` to probe its top 1,000 ports, reports open ones, and prints a summary table with device counts and vendor info.

## Troubleshooting

- **`pip install -e .` complains about `setup.py`:** Upgrade pip (`python -m pip install --upgrade pip`) or use `pip install .` if you do not need editable installs.
- **`uv pip install -e .` says "No virtual environment found":** Run `uv venv` first or re-run the install with `uv pip install --system -e .`.
- **CLI exits because `nmap`/`arp` is missing:** Install `nmap` via Homebrew and rerun. Use `netscanner --no-check` to skip the pre-flight validation if you already know the tools are present.
- **Vendor lookup fails:** The CLI will continue, but you can re-run later or disable lookups by working offline.

## Development

```bash
python -m pip install --upgrade pip
pip install -e .[dev]
ruff check src
pytest
```

Feel free to open issues or submit PRs for additional scanners, UI tweaks, or quality-of-life improvements.

## License

MIT License - see [LICENSE](LICENSE).
