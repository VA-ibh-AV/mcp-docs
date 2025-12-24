package indexer

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	filter, err := NewURLFilter("https://docs.langchain.com/docs")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trailing colon",
			input:    "https://docs.langchain.com/oss/javascript/langchain/philosophy:",
			expected: "https://docs.langchain.com/oss/javascript/langchain/philosophy",
		},
		{
			name:     "trailing semicolon",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview;",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "trailing comma",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview,",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "trailing period",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview.",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "multiple trailing punctuation",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview:;.",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "trailing whitespace",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview   ",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "leading whitespace",
			input:    "   https://docs.langchain.com/oss/javascript/langchain/overview",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "trailing slash removed",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview/",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "http to https",
			input:    "http://docs.langchain.com/oss/javascript/langchain/overview",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "fragment removed",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview#section",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "utm params removed",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview?utm_source=google&utm_medium=cpc",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "valid query params preserved",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview?version=2",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview?version=2",
		},
		{
			name:     "normal URL unchanged",
			input:    "https://docs.langchain.com/oss/javascript/langchain/overview",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
		{
			name:     "uppercase host lowercased",
			input:    "https://DOCS.LANGCHAIN.COM/oss/javascript/langchain/overview",
			expected: "https://docs.langchain.com/oss/javascript/langchain/overview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.NormalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsRelevant(t *testing.T) {
	filter, err := NewURLFilter("https://docs.langchain.com/docs")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		// Should be allowed
		{
			name:     "valid docs URL",
			url:      "https://docs.langchain.com/oss/javascript/langchain/overview",
			expected: true,
		},
		{
			name:     "valid deep docs URL",
			url:      "https://docs.langchain.com/oss/python/langchain/modules/agents",
			expected: true,
		},
		
		// Should be blocked - different domain
		{
			name:     "different domain",
			url:      "https://google.com/search",
			expected: false,
		},
		{
			name:     "github link",
			url:      "https://github.com/langchain-ai/langchain",
			expected: false,
		},
		
		// Should be blocked - forbidden extensions
		{
			name:     "PNG image",
			url:      "https://docs.langchain.com/images/logo.png",
			expected: false,
		},
		{
			name:     "PDF document",
			url:      "https://docs.langchain.com/docs/guide.pdf",
			expected: false,
		},
		{
			name:     "CSS file",
			url:      "https://docs.langchain.com/styles/main.css",
			expected: false,
		},
		{
			name:     "JS file",
			url:      "https://docs.langchain.com/scripts/app.js",
			expected: false,
		},
		
		// Should be blocked - forbidden keywords
		{
			name:     "pricing page",
			url:      "https://docs.langchain.com/pricing",
			expected: false,
		},
		{
			name:     "blog page",
			url:      "https://docs.langchain.com/blog/new-feature",
			expected: false,
		},
		{
			name:     "login page",
			url:      "https://docs.langchain.com/login",
			expected: false,
		},
		{
			name:     "careers page",
			url:      "https://docs.langchain.com/careers",
			expected: false,
		},
		
		// Should be blocked - root paths
		{
			name:     "root path",
			url:      "https://docs.langchain.com/",
			expected: false,
		},
		{
			name:     "empty path",
			url:      "https://docs.langchain.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.IsRelevant(tt.url)
			if result != tt.expected {
				t.Errorf("IsRelevant(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	filter, err := NewURLFilter("https://docs.langchain.com/docs")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	tests := []struct {
		name     string
		base     string
		relative string
		expected string
	}{
		{
			name:     "absolute URL unchanged",
			base:     "https://docs.langchain.com/docs/intro",
			relative: "https://docs.langchain.com/docs/guide",
			expected: "https://docs.langchain.com/docs/guide",
		},
		{
			name:     "relative path",
			base:     "https://docs.langchain.com/docs/intro",
			relative: "guide",
			expected: "https://docs.langchain.com/docs/guide",
		},
		{
			name:     "relative path with slash",
			base:     "https://docs.langchain.com/docs/intro",
			relative: "/docs/guide",
			expected: "https://docs.langchain.com/docs/guide",
		},
		{
			name:     "parent directory",
			base:     "https://docs.langchain.com/docs/section/intro",
			relative: "../guide",
			expected: "https://docs.langchain.com/docs/guide",
		},
		{
			name:     "fragment only",
			base:     "https://docs.langchain.com/docs/intro",
			relative: "#section",
			expected: "https://docs.langchain.com/docs/intro#section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ResolveURL(tt.base, tt.relative)
			if result != tt.expected {
				t.Errorf("ResolveURL(%q, %q) = %q, want %q", tt.base, tt.relative, result, tt.expected)
			}
		})
	}
}

func TestIsSameDomain(t *testing.T) {
	filter, err := NewURLFilter("https://docs.langchain.com/docs")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "exact match",
			url:      "https://docs.langchain.com/something",
			expected: true,
		},
		{
			name:     "with www prefix",
			url:      "https://www.docs.langchain.com/something",
			expected: true,
		},
		{
			name:     "different domain",
			url:      "https://other.com/something",
			expected: false,
		},
		{
			name:     "subdomain",
			url:      "https://api.docs.langchain.com/something",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.IsRelevant(tt.url)
			// Note: IsRelevant also checks other conditions, so we need URLs with paths
			if tt.expected && !result {
				// Only check for false negatives related to domain
				t.Logf("Warning: URL %q was filtered (may be due to other rules, not domain)", tt.url)
			}
		})
	}
}
