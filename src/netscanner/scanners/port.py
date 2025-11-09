"""Port scanning using nmap"""

import asyncio
import re
from typing import List
from rich.console import Console

from ..config import NMAP_TIMING, NMAP_MIN_RATE

console = Console()


async def scan_ports(ip: str, ports: List[int] = None) -> List[int]:
    """
    Scan ports on a device using nmap

    Args:
        ip: IP address to scan
        ports: List of ports to check (None = use nmap's top 1000 ports)

    Returns:
        List of open ports found
    """
    try:
        # Build nmap command
        nmap_args = ['nmap']

        if ports:
            # Use specific ports if provided
            port_str = ','.join(map(str, ports))
            nmap_args.extend(['-p', port_str])
        else:
            # Use nmap's default top 1000 most scanned ports
            nmap_args.append('--top-ports')
            nmap_args.append('1000')

        # Add common options
        nmap_args.extend([
            '--open',  # Only show open ports
            f'-{NMAP_TIMING}',  # Faster timing
            '--min-rate', NMAP_MIN_RATE,
            ip
        ])

        # Run nmap asynchronously
        process = await asyncio.create_subprocess_exec(
            *nmap_args,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )

        stdout, stderr = await process.communicate()

        open_ports = []
        for line in stdout.decode().split('\n'):
            # Look for lines like: "80/tcp open http"
            match = re.search(r'^(\d+)/tcp\s+open', line)
            if match:
                open_ports.append(int(match.group(1)))

        return open_ports

    except FileNotFoundError:
        console.print(f"[red]Error: nmap not found. Please install nmap to scan ports.[/red]")
        return []
    except Exception as e:
        console.print(f"[yellow]Warning: Could not scan {ip}: {e}[/yellow]")
        return []
