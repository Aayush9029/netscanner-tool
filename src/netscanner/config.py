"""Configuration and constants for network scanner"""

# Version
__version__ = "0.2.0"

# Scanning configuration
DEFAULT_CONCURRENCY = 5  # Max concurrent device scans
DEFAULT_TIMEOUT = 5  # Timeout in seconds
VENDOR_API_DELAY = 0.5  # Delay between vendor API calls to avoid rate limiting
VENDOR_API_URL = "https://api.macvendors.com/{mac}"

# Nmap configuration
NMAP_TIMING = "T4"  # Faster timing template
NMAP_MIN_RATE = "1000"  # Minimum packet rate
