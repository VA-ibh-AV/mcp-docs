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

	// Extract links
	links, err := bm.extractLinks(page)
	if err != nil {
		log.Printf("Warning: failed to extract links from %s: %v", url, err)
		links = []string{}
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

// extractLinks extracts all href links from the page
func (bm *BrowserManager) extractLinks(page playwright.Page) ([]string, error) {
	// JavaScript to extract all links
	result, err := page.Evaluate(`() => {
		const anchors = Array.from(document.querySelectorAll('a[href]'));
		return anchors.map(a => a.href).filter(href => href && href.startsWith('http'));
	}`)
	if err != nil {
		return nil, err
	}

	// Convert result to string slice
	linksInterface, ok := result.([]interface{})
	if !ok {
		return []string{}, nil
	}

	links := make([]string, 0, len(linksInterface))
	for _, link := range linksInterface {
		if linkStr, ok := link.(string); ok {
			links = append(links, linkStr)
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
	Links          []string
	StatusCode     int
	ResponseTimeMs int64
}
