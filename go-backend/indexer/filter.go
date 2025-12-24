package indexer

import (
	"net/url"
	"path"
	"strings"
)

// URLFilter determines which URLs should be crawled
type URLFilter struct {
	allowedDomain string
	baseURL       string
	basePath      string
}

// NewURLFilter creates a new URL filter for the given base URL
func NewURLFilter(baseURL string) (*URLFilter, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &URLFilter{
		allowedDomain: strings.ToLower(parsed.Hostname()),
		baseURL:       baseURL,
		basePath:      parsed.Path,
	}, nil
}

// Forbidden file extensions (assets, media, etc.)
var ForbiddenExtensions = map[string]bool{
	// Images
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".svg": true, ".webp": true, ".ico": true, ".bmp": true,
	".tiff": true, ".tif": true,

	// Stylesheets and scripts
	".css": true, ".js": true, ".mjs": true, ".map": true,
	".scss": true, ".sass": true, ".less": true,

	// Documents (non-HTML)
	".pdf": true, ".doc": true, ".docx": true, ".xls": true,
	".xlsx": true, ".ppt": true, ".pptx": true, ".odt": true,

	// Archives
	".zip": true, ".tar": true, ".gz": true, ".7z": true,
	".rar": true, ".bz2": true, ".xz": true,

	// Media
	".mp4": true, ".mp3": true, ".wav": true, ".avi": true,
	".mov": true, ".webm": true, ".ogg": true, ".flac": true,
	".mkv": true, ".wmv": true,

	// Fonts
	".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
	".otf": true,

	// Data files
	".json": true, ".xml": true, ".csv": true, ".yaml": true,
	".yml": true, ".toml": true,

	// Executables
	".exe": true, ".msi": true, ".dmg": true, ".deb": true,
	".rpm": true, ".apk": true, ".ipa": true,
}

// Forbidden URL path keywords (non-documentation content)
var ForbiddenKeywords = []string{
	// Commercial / Marketing
	"pricing", "plans", "enterprise", "customers", "case-study",
	"case-studies", "testimonials", "demo", "free-trial", "quote",

	// Company pages
	"about", "about-us", "company", "team", "careers", "jobs",
	"hiring", "work-with-us", "our-story", "leadership",

	// Legal / Compliance
	"security", "legal", "privacy", "privacy-policy", "terms",
	"terms-of-service", "tos", "gdpr", "cookies", "cookie-policy",
	"compliance", "dmca", "disclaimer",

	// Blog / News / Marketing content
	"blog", "news", "newsletter", "community", "event", "events",
	"press", "press-release", "updates", "release-notes", "releases",
	"changelog", "roadmap", "announcements", "webinar", "webinars",
	"podcast", "podcasts", "video", "videos",

	// Commerce
	"store", "shop", "buy", "purchase", "subscribe", "credits",
	"cart", "checkout", "billing", "payment", "upgrade",

	// Support (usually not documentation)
	"contact", "contact-us", "support", "help-center", "help-desk",
	"faq", "faqs", "ticket", "tickets", "feedback",

	// External code hosting (we want docs, not repo links)
	"github.com", "gitlab.com", "bitbucket.org", "codeberg.org",
	"sourceforge.net",

	// Auth pages
	"login", "logout", "signin", "sign-in", "signout", "sign-out",
	"signup", "sign-up", "register", "auth", "oauth", "sso",
	"password", "reset-password", "forgot-password", "verify",
	"confirm", "activate", "invitation",

	// User account pages
	"account", "profile", "settings", "preferences", "dashboard",
	"admin", "console", "portal",

	// Social / Sharing
	"share", "tweet", "facebook", "twitter", "linkedin", "reddit",
	"social", "follow",

	// Tracking / Analytics params (usually in query strings)
	"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
	"ref=", "source=", "campaign=", "affiliate=",

	// Download / Export
	"print-pdf", "pdf-version", "download", "export", "print",

	// Misc non-content
	"sitemap", "rss", "feed", "atom", "api/", "graphql", "rest/",
	"status", "health", "ping", "metrics", "telemetry",
}

// ForbiddenExactPaths are paths that should never be crawled
var ForbiddenExactPaths = map[string]bool{
	"/":           true, // Root page often not documentation
	"/index.html": true,
	"/index.htm":  true,
}

// IsRelevant checks if a URL should be crawled
func (f *URLFilter) IsRelevant(rawURL string) bool {
	// Parse the URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Normalize
	hostname := strings.ToLower(parsed.Hostname())
	urlPath := strings.ToLower(parsed.Path)

	// Must be same domain (with or without www)
	if !f.isSameDomain(hostname) {
		return false
	}

	// Check forbidden extensions
	ext := strings.ToLower(path.Ext(urlPath))
	if ext != "" && ForbiddenExtensions[ext] {
		return false
	}

	// Check exact forbidden paths
	if ForbiddenExactPaths[urlPath] {
		return false
	}

	// Check forbidden keywords in path and query
	fullURL := strings.ToLower(rawURL)
	for _, keyword := range ForbiddenKeywords {
		if strings.Contains(fullURL, keyword) {
			return false
		}
	}

	// Must have a path (not just domain)
	if urlPath == "" || urlPath == "/" {
		return false
	}

	return true
}

// isSameDomain checks if hostname matches allowed domain
func (f *URLFilter) isSameDomain(hostname string) bool {
	// Exact match
	if hostname == f.allowedDomain {
		return true
	}

	// With www prefix
	if hostname == "www."+f.allowedDomain {
		return true
	}

	// Subdomain match (e.g., docs.example.com for example.com)
	if strings.HasSuffix(hostname, "."+f.allowedDomain) {
		return true
	}

	return false
}

// NormalizeURL normalizes a URL for deduplication
func (f *URLFilter) NormalizeURL(rawURL string) string {
	// Trim whitespace and trailing punctuation that shouldn't be in URLs
	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.TrimRight(rawURL, ":;,.")
	
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove fragment
	parsed.Fragment = ""

	// Normalize scheme to https
	if parsed.Scheme == "http" {
		parsed.Scheme = "https"
	}

	// Remove trailing slash from path (unless it's root)
	if len(parsed.Path) > 1 && strings.HasSuffix(parsed.Path, "/") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/")
	}
	
	// Remove trailing colon from path (malformed URLs)
	parsed.Path = strings.TrimRight(parsed.Path, ":")

	// Remove common tracking parameters
	query := parsed.Query()
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"ref", "source", "campaign", "affiliate", "gclid", "fbclid",
	}
	for _, param := range trackingParams {
		query.Del(param)
	}
	parsed.RawQuery = query.Encode()

	// Lowercase hostname
	parsed.Host = strings.ToLower(parsed.Host)

	return parsed.String()
}

// ResolveURL resolves a relative URL against the base URL
func (f *URLFilter) ResolveURL(base, relative string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return relative
	}

	relURL, err := url.Parse(relative)
	if err != nil {
		return relative
	}

	return baseURL.ResolveReference(relURL).String()
}

// GetDomain returns the allowed domain
func (f *URLFilter) GetDomain() string {
	return f.allowedDomain
}
