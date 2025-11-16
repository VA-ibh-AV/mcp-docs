"""Simple CLI for building docs index using the project's indexer, embeddings and vectorstore.

This replaces the previous crew-based entrypoint and directly calls
`indexer.index_documentation` using an EmbeddingProvider implementation.
"""

import sys
import os

# Ensure the package src/ is on the import path so relative imports work.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from dotenv import load_dotenv
load_dotenv()

# Test scraper individually
def test_scraper():
    """Test the scraper function independently."""
    from indexer.scrapper import scrape_site
    from playwright.sync_api import sync_playwright, TimeoutError as PlaywrightTimeoutError
    from urllib.parse import urlparse
    from src.utils.url_filter import is_relevant_docs_url
    
    # Get URL from command line or use default
    test_url = sys.argv[2] if len(sys.argv) > 2 else "https://platform.openai.com/docs"
    
    print("=" * 70)
    print("SCRAPER TEST")
    print("=" * 70)
    print(f"\nTesting URL: {test_url}")
    print(f"{'='*70}\n")
    
    # First, let's manually inspect the first page to see what links exist
    print("Step 1: Fetching and analyzing the base page with Playwright...")
    try:
        with sync_playwright() as p:
            browser = p.chromium.launch(headless=True)
            context = browser.new_context(
                viewport={'width': 1920, 'height': 1080},
                user_agent='Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
            )
            page = context.new_page()
            page.set_default_timeout(30000)
            
            page.goto(test_url, wait_until="networkidle", timeout=30000)
            page.wait_for_timeout(2000)  # Wait for lazy content
            
            html = page.content()
            text = page.inner_text("body")
            
            print(f"✓ Successfully loaded page")
            print(f"  HTML length: {len(html)} characters")
            print(f"  Text length: {len(text)} characters")
            
            # Find all links using JavaScript
            all_links = page.evaluate("""
                () => {
                    const anchors = Array.from(document.querySelectorAll('a[href]'));
                    return anchors.map(a => a.href);
                }
            """)
            
            print(f"\n  Found {len(all_links)} total <a> tags with href attributes")
            
            # Analyze links
            domain = urlparse(test_url).netloc
            same_domain = []
            different_domain = []
            
            for link in all_links:
                parsed = urlparse(link)
                if parsed.netloc == domain or parsed.netloc == "":
                    same_domain.append(link)
                else:
                    different_domain.append(link)
            
            print(f"  Same domain links: {len(same_domain)}")
            print(f"  Different domain links: {len(different_domain)}")
            
            # Check which same-domain links pass the filter
            filtered_out = []
            passed_filter = []
            
            for link in same_domain[:50]:  # Show first 50
                parsed = urlparse(link)
                normalized = f"{parsed.scheme}://{parsed.netloc}{parsed.path}"
                if parsed.query:
                    normalized += f"?{parsed.query}"
                
                if is_relevant_docs_url(normalized, test_url):
                    passed_filter.append(normalized)
                else:
                    filtered_out.append(normalized)
            
            print(f"\n  Sample links analysis (first 50 same-domain):")
            print(f"    ✓ Passed filter: {len(passed_filter)}")
            if passed_filter:
                print(f"    Examples:")
                for link in passed_filter[:10]:
                    print(f"      - {link}")
            
            print(f"    ✗ Filtered out: {len(filtered_out)}")
            if filtered_out:
                print(f"    Examples:")
                for link in filtered_out[:10]:
                    print(f"      - {link}")
            
            browser.close()
        
    except Exception as e:
        print(f"❌ Error analyzing page: {e}")
        import traceback
        traceback.print_exc()
        return 1
    
    # Now run the actual scraper
    print(f"\n{'='*70}")
    print("Step 2: Running scraper...")
    print(f"{'='*70}\n")
    
    try:
        pages = scrape_site(test_url, max_pages=10, max_depth=2)
        
        print(f"\n{'='*70}")
        print(f"FINAL RESULTS")
        print(f"{'='*70}")
        print(f"Total pages scraped: {len(pages)}")
        
        if pages:
            for i, page in enumerate(pages, 1):
                print(f"\n--- Page {i} ---")
                print(f"URL: {page['url']}")
                print(f"Text length: {len(page['text'])} characters")
                preview = page['text'][:300].replace('\n', ' ').strip()
                print(f"Preview: {preview}...")
        else:
            print("\n⚠ No pages were scraped!")
            print("This could mean:")
            print("  - No links were found on the page")
            print("  - All links were filtered out")
            print("  - The page is JavaScript-rendered (SPA)")
            print("  - The URL filter is too restrictive")
                
    except Exception as e:
        print(f"❌ Error running scraper: {e}")
        import traceback
        traceback.print_exc()
        return 1
    
    return 0


def main():
    # Test scraper first
    if len(sys.argv) > 1 and sys.argv[1] == "test-scraper":
        return test_scraper()
    
    # Original indexing functionality
    from indexer import index_documentation
    from embeddings.openai_provider import OpenAIEmbeddingProvider
    
    url = os.environ.get("DOCS_URL", "https://docs.crewai.com/en/introduction")

    if OpenAIEmbeddingProvider is None:
        print("OpenAI embedding provider is not available. Install 'openai' package or check imports.")
        return 2

    try:
        provider = OpenAIEmbeddingProvider(api_key=os.environ.get("OPENAI_API_KEY"))
    except ValueError as e:
        print(f"OpenAI provider initialization failed: {e}")
        print("Set OPENAI_API_KEY environment variable or provide a valid api_key.")
        return 3

    print(f"Starting documentation indexing for: {url}")
    print("-" * 50)

    collection = index_documentation(url, provider,max_pages=100)

    print("-" * 50)
    if collection is not None:
        print("Indexing completed!")
        # Run a sample query against the newly-built index
        try:
            query_text = "how to create a crew"
            print(f"\nRunning sample query: '{query_text}'")
            q_vecs = provider.embed([query_text])
            if not q_vecs:
                print("No embedding returned for query; skipping search.")
            else:
                q_vec = q_vecs[0]
                results = collection.query(query_embedding=q_vec, n_results=5)

                # Chroma returns lists-of-lists when querying multiple queries.
                # Normalize fields to lists of results for the single query.
                def _first_or_list(x):
                    if x is None:
                        return []
                    if isinstance(x, list) and len(x) > 0 and isinstance(x[0], list):
                        return x[0]
                    return x

                ids = _first_or_list(results.get("ids"))
                docs = _first_or_list(results.get("documents"))
                dists = _first_or_list(results.get("distances"))
                metas = _first_or_list(results.get("metadatas"))

                if not docs:
                    print("No documents returned for the query.")
                else:
                    for i, doc in enumerate(docs):
                        meta = metas[i] if metas and i < len(metas) else {}
                        dist = dists[i] if dists and i < len(dists) else None
                        idv = ids[i] if ids and i < len(ids) else None
                        print("\n---")
                        print(f"Rank {i+1} (id={idv}, distance={dist}):")
                        print(doc[:1000])
                        if meta:
                            print(f"metadata: {meta}")
        except Exception as e:
            print(f"Sample query failed: {e}")
    else:
        print("Indexing finished with no results.")

    return 0


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "test-scraper":
        exit_code = test_scraper()
    else:
        exit_code = main()
    sys.exit(exit_code)
