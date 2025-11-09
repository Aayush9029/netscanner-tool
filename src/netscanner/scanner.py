"""Main network scanner class"""

import asyncio
from typing import List
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn

from .models import Device
from .config import DEFAULT_CONCURRENCY, VENDOR_API_DELAY
from .scanners import get_arp_table, scan_ports, lookup_mac_vendor

console = Console()


class NetworkScanner:
    """Async network scanner using ARP and nmap"""

    def __init__(self, ports: List[int] = None, max_concurrent: int = DEFAULT_CONCURRENCY):
        """
        Initialize the network scanner

        Args:
            ports: List of ports to scan (None = nmap's top 1000 ports)
            max_concurrent: Maximum number of concurrent device scans
        """
        self.ports = ports
        self.max_concurrent = max_concurrent
        self.devices: List[Device] = []

    async def scan_device(self, device: Device, progress=None, task_id=None) -> Device:
        """
        Scan a single device for open ports and vendor info

        Args:
            device: Device to scan
            progress: Optional progress bar
            task_id: Optional progress task ID

        Returns:
            Scanned device with updated information
        """
        # Look up vendor
        device.vendor = await lookup_mac_vendor(device.mac)

        # Small delay to avoid rate limiting
        await asyncio.sleep(VENDOR_API_DELAY)

        # Scan ports
        device.open_ports = await scan_ports(device.ip, self.ports)

        if progress and task_id:
            progress.update(task_id, advance=1)

        return device

    async def scan_network(self) -> List[Device]:
        """
        Main scanning function

        Returns:
            List of scanned devices
        """
        # Get devices from ARP
        devices = await get_arp_table()

        if not devices:
            console.print("[yellow]No devices found in ARP table[/yellow]")
            return []

        # Scan all devices concurrently
        console.print(f"[cyan]Scanning {len(devices)} devices for open ports and vendor info...[/cyan]")

        with Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        ) as progress:
            task = progress.add_task("Scanning devices...", total=len(devices))

            # Scan devices with concurrency limit to avoid overwhelming the network
            semaphore = asyncio.Semaphore(self.max_concurrent)

            async def scan_with_semaphore(device):
                async with semaphore:
                    return await self.scan_device(device, progress, task)

            self.devices = await asyncio.gather(*[
                scan_with_semaphore(device) for device in devices
            ])

        return self.devices
