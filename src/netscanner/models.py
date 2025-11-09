"""Data models for network scanner"""

from dataclasses import dataclass, field
from typing import List, Optional


@dataclass
class Device:
    """Represents a network device"""
    ip: str
    mac: str
    vendor: Optional[str] = None
    hostname: Optional[str] = None
    open_ports: List[int] = field(default_factory=list)

    def __str__(self) -> str:
        """String representation of device"""
        ports = ', '.join(map(str, self.open_ports)) if self.open_ports else "None"
        return f"{self.ip} ({self.mac}) - {self.vendor or 'Unknown'} - Ports: {ports}"
