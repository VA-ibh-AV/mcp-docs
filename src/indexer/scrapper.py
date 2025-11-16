# src/docs_mcp/indexer/scraper.py

try:
    from playwright.sync_api import sync_playwright, TimeoutError as PlaywrightTimeoutError
except ImportError as e:
    raise RuntimeError(
        "Missing dependency 'playwright'. Install with: `pip install playwright` and then `playwright install chromium`"
    ) from e

from urllib.parse import urljoin, urlparse
from src.utils.url_filter import is_relevant_docs_url


def scrape_site(base_url, max_pages=10, max_depth=3):
    """
    Scrape a documentation site using Playwright to handle JavaScript-rendered pages.
    
    Args:
        base_url: Starting URL to scrape
        max_pages: Maximum number of pages to scrape
        max_depth: Maximum depth to crawl (0 = only base URL)
    
    Returns:
        List of dicts with 'url', 'html', and 'text' keys
    """
    visited = set()
    pages = []
    queue = [(base_url, 0)]
    domain = urlparse(base_url).netloc
    links_found = 0
    links_filtered = 0
    links_added = 0

    with sync_playwright() as p:
        # Launch browser in headless mode
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            viewport={'width': 1920, 'height': 1080},
            user_agent='Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
        )
        page = context.new_page()
        
        # Set a reasonable timeout
        page.set_default_timeout(30000)  # 30 seconds
        
        while queue:
            url, depth = queue.pop(0)

            if depth > max_depth:
                continue
            if url in visited:
                continue
            if len(pages) >= max_pages:
                break

            visited.add(url)

            # Always scrape the base URL (depth 0), but filter others
            should_scrape = (depth == 0) or is_relevant_docs_url(url, base_url)
            
            if not should_scrape:
                continue
            
            print(f"Scraping: {url} (depth: {depth})")
            try:
                # Navigate to the page and wait for it to load
                page.goto(url, wait_until="networkidle", timeout=30000)
                
                # Wait a bit more for any lazy-loaded content
                page.wait_for_timeout(2000)  # 2 seconds
                
                # Get the rendered HTML and text
                html = page.content()
                text = page.inner_text("body")
                
            except PlaywrightTimeoutError as e:
                print(f"  ⚠ Timeout loading {url}: {e}")
                continue
            except Exception as e:
                print(f"  ⚠ Failed to fetch {url}: {e}")
                continue

            pages.append({
                "url": url,
                "html": html,
                "text": text
            })
            print(f"  ✓ Scraped {len(text)} characters")

            # Extract links from the rendered page
            try:
                # Get all anchor tags with href attributes
                links = page.evaluate("""
                    () => {
                        const anchors = Array.from(document.querySelectorAll('a[href]'));
                        return anchors.map(a => a.href);
                    }
                """)
                
                for link_url in links:
                    if not link_url:
                        continue
                    
                    links_found += 1
                    
                    # Normalize URL: remove fragment, handle query params
                    parsed = urlparse(link_url)
                    # Remove fragment but keep query params
                    normalized = f"{parsed.scheme}://{parsed.netloc}{parsed.path}"
                    if parsed.query:
                        normalized += f"?{parsed.query}"
                    
                    # Only process same-domain links
                    parsed_normalized = urlparse(normalized)
                    if parsed_normalized.netloc != domain and parsed_normalized.netloc:
                        links_filtered += 1
                        continue
                    
                    # Check if already visited
                    if normalized in visited:
                        continue
                    
                    # Filter links before adding to queue (except we always process base URL)
                    if depth == 0 or is_relevant_docs_url(normalized, base_url):
                        if normalized not in [q[0] for q in queue]:
                            queue.append((normalized, depth + 1))
                            links_added += 1
                    else:
                        links_filtered += 1
                        
            except Exception as e:
                print(f"  ⚠ Error extracting links from {url}: {e}")
                continue

        browser.close()

    print(f"\n📊 Scraping summary:")
    print(f"   Pages scraped: {len(pages)}")
    print(f"   Links found: {links_found}")
    print(f"   Links added to queue: {links_added}")
    print(f"   Links filtered: {links_filtered}")
    print(f"   URLs visited: {len(visited)}")
    
    return pages
