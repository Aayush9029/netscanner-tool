"""Command-line interface for network scanner"""

import asyncio
import click
from rich.console import Console

from .config import DEFAULT_CONCURRENCY, __version__
from .scanner import NetworkScanner
from .display import display_banner, display_results, display_summary
from .utils.system_check import verify_system_dependencies

console = Console()


@click.command()
@click.option(
    '--ports',
    '-p',
    help='Comma-separated list of ports to scan (e.g., "22,80,443")',
    default=None,
)
@click.option(
    '--max-concurrent',
    '-m',
    type=int,
    default=DEFAULT_CONCURRENCY,
    help=f'Maximum number of concurrent device scans (default: {DEFAULT_CONCURRENCY})',
)
@click.option(
    '--no-check',
    is_flag=True,
    help='Skip system dependency check',
)
@click.version_option(version=__version__, prog_name='netscanner')
def main(ports: str, max_concurrent: int, no_check: bool) -> None:
    """
    Network Scanner - Discover devices on your local network

    This tool uses ARP tables and nmap to find devices and their open ports.
    """
    try:
        # Check system dependencies unless skipped
        if not no_check:
            verify_system_dependencies(strict=True)

        # Parse custom ports if provided
        port_list = None
        if ports:
            try:
                port_list = [int(p.strip()) for p in ports.split(',')]
                console.print(f"[cyan]Scanning custom ports: {port_list}[/cyan]")
            except ValueError:
                console.print("[red]Error: Invalid port list. Use comma-separated numbers.[/red]")
                return

        # Display banner
        display_banner()

        # Run the scan
        scanner = NetworkScanner(ports=port_list, max_concurrent=max_concurrent)
        devices = asyncio.run(scanner.scan_network())

        # Display results
        if devices:
            display_results(devices)
            display_summary(devices)
        else:
            console.print("[yellow]No devices found.[/yellow]")

    except KeyboardInterrupt:
        console.print("\n[yellow]Scan cancelled by user[/yellow]")
    except Exception as e:
        console.print(f"\n[red]Error: {e}[/red]")
        raise


if __name__ == "__main__":
    main()
