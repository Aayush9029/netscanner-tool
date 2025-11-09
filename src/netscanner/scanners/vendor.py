"""MAC vendor lookup via API"""

import asyncio
from typing import Optional
import aiohttp

from ..config import VENDOR_API_URL, DEFAULT_TIMEOUT, VENDOR_API_DELAY


async def lookup_mac_vendor(mac: str) -> Optional[str]:
    """
    Look up MAC address vendor using macvendors.com API

    Args:
        mac: MAC address to lookup

    Returns:
        Vendor name if found, None otherwise
    """
    try:
        async with aiohttp.ClientSession() as session:
            url = VENDOR_API_URL.format(mac=mac)
            async with session.get(
                url,
                timeout=aiohttp.ClientTimeout(total=DEFAULT_TIMEOUT)
            ) as response:
                if response.status == 200:
                    vendor = await response.text()
                    return vendor.strip()
                else:
                    return None
    except asyncio.TimeoutError:
        return None
    except Exception:
        return None


async def lookup_vendors_with_delay(macs: list[str]) -> dict[str, Optional[str]]:
    """
    Look up multiple MAC vendors with rate limiting

    Args:
        macs: List of MAC addresses

    Returns:
        Dictionary mapping MAC to vendor name
    """
    results = {}
    for mac in macs:
        results[mac] = await lookup_mac_vendor(mac)
        # Add delay to avoid rate limiting
        await asyncio.sleep(VENDOR_API_DELAY)
    return results
