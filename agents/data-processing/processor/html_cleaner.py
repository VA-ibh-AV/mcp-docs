"""
HTML content decompression and cleaning utilities.
"""

import base64
import gzip
import logging
import re
from io import BytesIO
from typing import Optional

from bs4 import BeautifulSoup

logger = logging.getLogger(__name__)


def decompress_html(compressed: str, encoding: str = "gzip+base64") -> str:
    """
    Decompress HTML content from gzip+base64 encoding.
    
    Args:
        compressed: Compressed HTML string
        encoding: Encoding type ("gzip+base64" or "plain")
        
    Returns:
        Decompressed HTML string
    """
    if encoding == "plain" or not compressed:
        return compressed
    
    if encoding == "gzip+base64":
        try:
            # Base64 decode
            compressed_bytes = base64.b64decode(compressed)
            
            # Gzip decompress
            with gzip.GzipFile(fileobj=BytesIO(compressed_bytes)) as gz:
                return gz.read().decode("utf-8")
                
        except Exception as e:
            logger.error(f"Failed to decompress HTML: {e}")
            raise ValueError(f"Failed to decompress HTML: {e}")
    
    raise ValueError(f"Unknown encoding: {encoding}")


def clean_html_to_text(html: str) -> str:
    """
    Extract clean text from HTML content.
    
    Removes scripts, styles, and extracts readable text.
    
    Args:
        html: Raw HTML string
        
    Returns:
        Clean text content
    """
    if not html:
        return ""
    
    try:
        soup = BeautifulSoup(html, "html.parser")
        
        # Remove script and style elements
        for element in soup(["script", "style", "nav", "footer", "header", "aside"]):
            element.decompose()
        
        # Get text
        text = soup.get_text(separator=" ", strip=True)
        
        # Clean up whitespace
        text = re.sub(r"\s+", " ", text)
        text = text.strip()
        
        return text
        
    except Exception as e:
        logger.error(f"Failed to clean HTML: {e}")
        return ""


def extract_title(html: str) -> Optional[str]:
    """
    Extract page title from HTML.
    
    Args:
        html: Raw HTML string
        
    Returns:
        Page title or None
    """
    if not html:
        return None
    
    try:
        soup = BeautifulSoup(html, "html.parser")
        title_tag = soup.find("title")
        if title_tag:
            return title_tag.get_text(strip=True)
        return None
    except Exception:
        return None
