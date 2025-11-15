# src/docs_mcp/indexer/scraper.py

try:
    import requests
except ImportError as e:
    raise RuntimeError(
        "Missing dependency 'requests'. Install project dependencies with: `pip install -r requirements.txt` "
        "or add 'requests' to your environment. Original ImportError: " + str(e)
    ) from e

from bs4 import BeautifulSoup
from urllib.parse import urljoin, urlparse
from utils.url_filter import is_relevant_docs_url


def scrape_site(base_url, max_pages=10, max_depth=3):
    visited = set()
    pages = []
    queue = [(base_url, 0)]
    domain = urlparse(base_url).netloc

    while queue:
        url, depth = queue.pop(0)

        if depth > max_depth:
            continue
        if url in visited:
            continue
        if len(pages) >= max_pages:
            break

        visited.add(url)

        if not is_relevant_docs_url(url, base_url):
            continue
        
        print(f"Scraping: {url} (depth: {depth})")
        try:
            res = requests.get(url, timeout=10)
        except:
            continue

        soup = BeautifulSoup(res.text, "html.parser")
        text = soup.get_text(" ", strip=True)

        pages.append({
            "url": url,
            "html": res.text,
            "text": text
        })

        # extract new links
        for a in soup.find_all("a"):
            href = a.get("href")
            if not href:
                continue

            full = urljoin(url, href)
            if urlparse(full).netloc == domain:
                queue.append((full, depth + 1))

    return pages
