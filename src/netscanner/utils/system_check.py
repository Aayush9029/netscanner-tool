"""System dependency verification"""

import shutil
import subprocess
import sys
from typing import Tuple, List
from rich.console import Console

console = Console()


class SystemCheckError(Exception):
    """Raised when system dependencies are missing"""
    pass


def check_command_exists(command: str) -> bool:
    """Check if a command exists in PATH"""
    return shutil.which(command) is not None


def check_nmap() -> Tuple[bool, str]:
    """
    Check if nmap is installed and get version

    Returns:
        Tuple of (success, message)
    """
    if not check_command_exists('nmap'):
        return False, "nmap not found"

    try:
        result = subprocess.run(
            ['nmap', '--version'],
            capture_output=True,
            text=True,
            timeout=5
        )
        version_line = result.stdout.split('\n')[0]
        return True, version_line
    except Exception as e:
        return False, f"Error checking nmap: {e}"


def check_arp() -> Tuple[bool, str]:
    """
    Check if arp command is available

    Returns:
        Tuple of (success, message)
    """
    if not check_command_exists('arp'):
        return False, "arp command not found"
    return True, "arp available"


def verify_system_dependencies(strict: bool = True) -> bool:
    """
    Verify all required system dependencies

    Args:
        strict: If True, exit on missing dependencies. If False, warn only.

    Returns:
        bool: True if all dependencies satisfied
    """
    missing_deps: List[str] = []

    console.print("\n[bold cyan]Checking system dependencies...[/bold cyan]")

    # Check nmap
    nmap_ok, nmap_msg = check_nmap()
    if nmap_ok:
        console.print(f"[green]✓[/green] {nmap_msg}")
    else:
        console.print(f"[red]✗[/red] nmap: {nmap_msg}")
        missing_deps.append("nmap")

    # Check arp
    arp_ok, arp_msg = check_arp()
    if arp_ok:
        console.print(f"[green]✓[/green] {arp_msg}")
    else:
        console.print(f"[red]✗[/red] arp: {arp_msg}")
        missing_deps.append("arp")

    if missing_deps:
        console.print("\n[bold yellow]Missing required dependencies:[/bold yellow]")
        for dep in missing_deps:
            console.print(f"  - {dep}")

        console.print("\n[bold]Installation:[/bold]")
        console.print("  [cyan]brew install nmap[/cyan]")
        console.print()

        if strict:
            console.print("[red]Please install missing dependencies and try again.[/red]\n")
            sys.exit(1)
        return False

    console.print("[green]All dependencies satisfied![/green]\n")
    return True
