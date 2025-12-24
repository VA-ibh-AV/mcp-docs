package indexer

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// BrowserManager handles Playwright browser lifecycle and page pooling
type BrowserManager struct {
	pw       *playwright.Playwright
	browser  playwright.Browser
	pagePool chan playwright.Page
	config   *Config
	mu       sync.Mutex
	closed   bool
}

// NewBrowserManager creates and initializes a new browser manager
func NewBrowserManager(config *Config) (*BrowserManager, error) {
	// Install playwright browsers if needed
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		log.Printf("Warning: playwright install returned error (may already be installed): %v", err)
	}

	// Start Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to start playwright: %w", err)
	}

	// Launch browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(config.Headless),
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Create page pool
	pagePool := make(chan playwright.Page, config.MaxConcurrency)

	bm := &BrowserManager{
		pw:       pw,
		browser:  browser,
		pagePool: pagePool,
		config:   config,
	}

	// Pre-create pages for the pool
	for i := 0; i < config.MaxConcurrency; i++ {
		page, err := bm.createPage()
		if err != nil {
			bm.Close()
			return nil, fmt.Errorf("failed to create page %d: %w", i, err)
		}
		pagePool <- page
	}

	return bm, nil
}

// createPage creates a new browser page with configured settings
func (bm *BrowserManager) createPage() (playwright.Page, error) {
	context, err := bm.browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String(bm.config.UserAgent),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		return nil, err
	}

	page, err := context.NewPage()
	if err != nil {
		context.Close()
		return nil, err
	}

	// Set default timeout
	page.SetDefaultTimeout(float64(bm.config.PageTimeout.Milliseconds()))

	return page, nil
}

// AcquirePage gets a page from the pool (blocks if none available)
func (bm *BrowserManager) AcquirePage() (playwright.Page, error) {
	bm.mu.Lock()
	if bm.closed {
		bm.mu.Unlock()
		return nil, fmt.Errorf("browser manager is closed")
	}
	bm.mu.Unlock()

	select {
	case page := <-bm.pagePool:
		return page, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for available page")
	}
}

// ReleasePage returns a page to the pool
func (bm *BrowserManager) ReleasePage(page playwright.Page) {
	bm.mu.Lock()
	if bm.closed {
		bm.mu.Unlock()
		page.Context().Close()
		return
	}
	bm.mu.Unlock()

	// Clear page state before returning to pool
	// Navigate to blank to clear cookies/storage
	page.Goto("about:blank")

	select {
	case bm.pagePool <- page:
		// Returned to pool
	default:
		// Pool is full, close the page
		page.Context().Close()
	}
}

// FetchPage navigates to a URL and extracts content
func (bm *BrowserManager) FetchPage(page playwright.Page, url string) (*PageFetchResult, error) {
	startTime := time.Now()

	// Navigate to the page
	response, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(bm.config.PageTimeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Wait additional time for JS rendering
	page.WaitForTimeout(float64(bm.config.IdleTimeout.Milliseconds()))

	// Get status code
	statusCode := 0
	if response != nil {
		statusCode = response.Status()
	}

	// Extract HTML content
	html, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("failed to get page content: %w", err)
	}

	// Extract text content
	text := ""
	if bm.config.ExtractText {
		textContent, err := page.InnerText("body")
		if err == nil {
			text = textContent
			if len(text) > bm.config.MaxTextLength {
				text = text[:bm.config.MaxTextLength]
			}
		}
	}

	// Extract title
	title, _ := page.Title()

	// Extract links with source classification
	links, err := bm.extractLinks(page)
	if err != nil {
		log.Printf("Warning: failed to extract links from %s: %v", url, err)
		links = []ExtractedLink{}
	}

	return &PageFetchResult{
		URL:            url,
		HTML:           html,
		Text:           text,
		Title:          title,
		Links:          links,
		StatusCode:     statusCode,
		ResponseTimeMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// extractLinks extracts all href links from the page with source classification
func (bm *BrowserManager) extractLinks(page playwright.Page) ([]ExtractedLink, error) {
	// JavaScript to extract all links with smart sidebar/footer detection
	result, err := page.Evaluate(`() => {
		const viewportWidth = window.innerWidth;
		const viewportHeight = window.innerHeight;
		const docHeight = Math.max(
			document.body.scrollHeight,
			document.documentElement.scrollHeight
		);
		
		// Pre-compute sidebar containers using heuristics
		const sidebarContainers = new Set();
		const footerContainers = new Set();
		
		// Analyze all potential containers
		const containers = document.querySelectorAll('nav, aside, div, section, ul, header, footer');
		
		for (const el of containers) {
			const rect = el.getBoundingClientRect();
			if (rect.width === 0 || rect.height === 0) continue;
			
			const style = window.getComputedStyle(el);
			const linkCount = el.querySelectorAll('a[href]').length;
			
			// Sidebar detection heuristics
			const isNarrow = rect.width > 0 && rect.width < 400;
			const isOnLeftEdge = rect.left < 100;
			const isOnRightEdge = rect.right > viewportWidth - 100;
			const isSticky = style.position === 'fixed' || style.position === 'sticky';
			const hasLinks = linkCount >= 5;
			const isTall = rect.height > viewportHeight * 0.5;
			const tagName = el.tagName.toLowerCase();
			
			// Score-based sidebar detection
			let sidebarScore = 0;
			if (isNarrow) sidebarScore += 3;
			if (isOnLeftEdge || isOnRightEdge) sidebarScore += 3;
			if (isSticky) sidebarScore += 2;
			if (hasLinks) sidebarScore += 2;
			if (isTall) sidebarScore += 1;
			if (tagName === 'nav' || tagName === 'aside') sidebarScore += 2;
			
			// Check for common sidebar class/id patterns
			const classAndId = (el.className + ' ' + el.id).toLowerCase();
			if (/sidebar|sidenav|side-nav|toc|menu|nav-menu|navigation|docs-nav/.test(classAndId)) {
				sidebarScore += 3;
			}
			
			if (sidebarScore >= 5) {
				sidebarContainers.add(el);
			}
			
			// Footer detection
			const isAtBottom = rect.top > docHeight - 300;
			const isFooterTag = tagName === 'footer';
			const hasFooterClass = /footer/.test(classAndId);
			
			if (isFooterTag || hasFooterClass || isAtBottom) {
				footerContainers.add(el);
			}
		}
		
		// Helper to check if element is inside a container set
		function isInsideContainerSet(element, containerSet) {
			let parent = element;
			while (parent) {
				if (containerSet.has(parent)) return true;
				parent = parent.parentElement;
			}
			return false;
		}
		
		// Extract and classify all links
		const links = document.querySelectorAll('a[href]');
		const results = [];
		const seen = new Set();
		
		for (const link of links) {
			if (!link.href || !link.href.startsWith('http')) continue;
			if (seen.has(link.href)) continue;
			seen.add(link.href);
			
			let source = 'content';
			
			// Check if inside sidebar
			if (isInsideContainerSet(link, sidebarContainers)) {
				source = 'sidebar';
			}
			// Check if inside footer
			else if (isInsideContainerSet(link, footerContainers)) {
				source = 'footer';
			}
			
			results.push({ url: link.href, source: source });
		}
		
		return results;
	}`)
	if err != nil {
		return nil, err
	}

	// Convert result to ExtractedLink slice
	linksInterface, ok := result.([]interface{})
	if !ok {
		return []ExtractedLink{}, nil
	}

	links := make([]ExtractedLink, 0, len(linksInterface))
	for _, item := range linksInterface {
		if linkMap, ok := item.(map[string]interface{}); ok {
			url, _ := linkMap["url"].(string)
			source, _ := linkMap["source"].(string)
			if url != "" {
				links = append(links, ExtractedLink{URL: url, Source: source})
			}
		}
	}

	return links, nil
}

// Close shuts down the browser manager
func (bm *BrowserManager) Close() error {
	bm.mu.Lock()
	if bm.closed {
		bm.mu.Unlock()
		return nil
	}
	bm.closed = true
	bm.mu.Unlock()

	// Close all pages in the pool
	close(bm.pagePool)
	for page := range bm.pagePool {
		page.Context().Close()
	}

	// Close browser and playwright
	if bm.browser != nil {
		bm.browser.Close()
	}
	if bm.pw != nil {
		bm.pw.Stop()
	}

	return nil
}

// PageFetchResult holds the result of fetching a page
type PageFetchResult struct {
	URL            string
	HTML           string
	Text           string
	Title          string
	Links          []ExtractedLink
	StatusCode     int
	ResponseTimeMs int64
}
