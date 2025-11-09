"""Display and UI functions using Rich"""

from typing import List
from rich.console import Console
from rich.table import Table
from rich.panel import Panel

from .models import Device

console = Console()


def display_banner() -> None:
    """Display the application banner"""
    console.print(Panel.fit(
        "[bold cyan]Network Scanner[/bold cyan]\n"
        "Discovering devices on your local network...",
        border_style="cyan"
    ))
    console.print()


def display_results(devices: List[Device]) -> None:
    """
    Display scan results in a nice table

    Args:
        devices: List of scanned devices
    """
    if not devices:
        console.print("[yellow]No devices to display[/yellow]")
        return

    table = Table(title="Network Devices", show_header=True, header_style="bold magenta")
    table.add_column("IP Address", style="cyan", no_wrap=True)
    table.add_column("MAC Address", style="green")
    table.add_column("Vendor", style="yellow")
    table.add_column("Hostname", style="blue")
    table.add_column("Open Ports", style="red")

    # Sort by IP address
    sorted_devices = sorted(devices, key=lambda d: [int(x) for x in d.ip.split('.')])

    for device in sorted_devices:
        ports_str = ', '.join(map(str, device.open_ports)) if device.open_ports else "None"
        vendor = device.vendor or "Unknown"
        hostname = device.hostname or "-"

        table.add_row(
            device.ip,
            device.mac,
            vendor,
            hostname,
            ports_str
        )

    console.print()
    console.print(table)
    console.print()


def display_summary(devices: List[Device]) -> None:
    """
    Display summary statistics

    Args:
        devices: List of scanned devices
    """
    total_devices = len(devices)
    devices_with_ports = len([d for d in devices if d.open_ports])
    total_open_ports = sum(len(d.open_ports) for d in devices)

    summary = Panel(
        f"[green]Total Devices:[/green] {total_devices}\n"
        f"[yellow]Devices with Open Ports:[/yellow] {devices_with_ports}\n"
        f"[red]Total Open Ports:[/red] {total_open_ports}",
        title="Summary",
        border_style="blue"
    )
    console.print(summary)
    console.print()


def display_error(message: str) -> None:
    """
    Display an error message

    Args:
        message: Error message to display
    """
    console.print(f"[red]Error: {message}[/red]")


def display_warning(message: str) -> None:
    """
    Display a warning message

    Args:
        message: Warning message to display
    """
    console.print(f"[yellow]Warning: {message}[/yellow]")


def display_info(message: str) -> None:
    """
    Display an info message

    Args:
        message: Info message to display
    """
    console.print(f"[cyan]{message}[/cyan]")
