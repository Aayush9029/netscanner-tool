"""ARP table parsing for device discovery"""

import re
import subprocess
from typing import List
from rich.console import Console

from ..models import Device

console = Console()


async def get_arp_table() -> List[Device]:
    """
    Parse ARP table to get IP and MAC addresses

    Returns:
        List of discovered devices from ARP table
    """
    console.print("[cyan]Scanning ARP table...[/cyan]")

    try:
        result = subprocess.run(
            ['arp', '-a'],
            capture_output=True,
            text=True,
            check=True,
            timeout=10
        )

        devices = []
        # Parse ARP output - format varies by OS
        # macOS: hostname (ip) at mac on interface
        # Linux: hostname (ip) at mac [ether] on interface

        for line in result.stdout.split('\n'):
            # macOS format: ? (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]
            mac_match = re.search(r'at ([0-9a-fA-F:]{17})', line)
            ip_match = re.search(r'\((\d+\.\d+\.\d+\.\d+)\)', line)
            hostname_match = re.search(r'^(\S+)\s+\(', line)

            if mac_match and ip_match:
                mac = mac_match.group(1).lower()
                ip = ip_match.group(1)
                hostname = hostname_match.group(1) if hostname_match and hostname_match.group(1) != '?' else None

                # Skip incomplete entries and broadcast addresses
                if 'incomplete' not in line.lower() and mac != 'ff:ff:ff:ff:ff:ff':
                    devices.append(Device(ip=ip, mac=mac, hostname=hostname))

        console.print(f"[green]Found {len(devices)} devices in ARP table[/green]")
        return devices

    except subprocess.CalledProcessError as e:
        console.print(f"[red]Error running arp command: {e}[/red]")
        return []
    except subprocess.TimeoutExpired:
        console.print("[red]ARP command timed out[/red]")
        return []
    except Exception as e:
        console.print(f"[red]Unexpected error in ARP scan: {e}[/red]")
        return []
