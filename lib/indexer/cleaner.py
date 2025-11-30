# src/docs_mcp/indexer/cleaner.py

import re

def clean_page_text(text: str) -> str:
    text = re.sub(r"\s+", " ", text)
    return text.strip()
