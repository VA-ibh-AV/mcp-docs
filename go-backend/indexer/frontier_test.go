package indexer

import (
	"testing"
)

func TestFrontierPriority(t *testing.T) {
	filter, err := NewURLFilter("https://docs.example.com")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	frontier := NewURLFrontier(filter, 5, 100)

	// Add links in mixed order: content, footer, sidebar
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/content1",
		Depth:    1,
		Priority: LinkPriorityContent,
		Source:   LinkSourceContent,
	})
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/footer1",
		Depth:    1,
		Priority: LinkPriorityFooter,
		Source:   LinkSourceFooter,
	})
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/sidebar1",
		Depth:    1,
		Priority: LinkPrioritySidebar,
		Source:   LinkSourceSidebar,
	})
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/sidebar2",
		Depth:    1,
		Priority: LinkPrioritySidebar,
		Source:   LinkSourceSidebar,
	})
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/content2",
		Depth:    1,
		Priority: LinkPriorityContent,
		Source:   LinkSourceContent,
	})

	// Pop items and verify priority groups are in order
	// Sidebar (priority 1) -> Content (priority 6) -> Footer (priority 11)
	// Within same priority, order is not guaranteed (heap property)
	
	// First 2 should be sidebar
	item1 := frontier.Pop()
	item2 := frontier.Pop()
	if item1 == nil || item2 == nil {
		t.Fatal("Expected 2 sidebar items")
	}
	if !contains(item1.URL, "sidebar") || !contains(item2.URL, "sidebar") {
		t.Errorf("First 2 items should be sidebar, got %s and %s", item1.URL, item2.URL)
	}
	
	// Next 2 should be content
	item3 := frontier.Pop()
	item4 := frontier.Pop()
	if item3 == nil || item4 == nil {
		t.Fatal("Expected 2 content items")
	}
	if !contains(item3.URL, "content") || !contains(item4.URL, "content") {
		t.Errorf("Next 2 items should be content, got %s and %s", item3.URL, item4.URL)
	}
	
	// Last should be footer
	item5 := frontier.Pop()
	if item5 == nil {
		t.Fatal("Expected 1 footer item")
	}
	if !contains(item5.URL, "footer") {
		t.Errorf("Last item should be footer, got %s", item5.URL)
	}

	// Queue should be empty
	if item := frontier.Pop(); item != nil {
		t.Errorf("Expected empty queue, got %s", item.URL)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFrontierPriorityWithDepth(t *testing.T) {
	filter, err := NewURLFilter("https://docs.example.com")
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}

	frontier := NewURLFrontier(filter, 5, 100)

	// Sidebar at depth 2 should still come before content at depth 1
	// Because priority = source_priority + depth
	// Sidebar depth 2: 0 + 2 = 2
	// Content depth 1: 5 + 1 = 6
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/content-d1",
		Depth:    1,
		Priority: LinkPriorityContent,
		Source:   LinkSourceContent,
	})
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/sidebar-d2",
		Depth:    2,
		Priority: LinkPrioritySidebar,
		Source:   LinkSourceSidebar,
	})
	frontier.Push(URLItem{
		URL:      "https://docs.example.com/sidebar-d1",
		Depth:    1,
		Priority: LinkPrioritySidebar,
		Source:   LinkSourceSidebar,
	})

	// Expected: sidebar-d1 (0+1=1), sidebar-d2 (0+2=2), content-d1 (5+1=6)
	expectedOrder := []string{
		"https://docs.example.com/sidebar-d1",
		"https://docs.example.com/sidebar-d2",
		"https://docs.example.com/content-d1",
	}

	for i, expected := range expectedOrder {
		item := frontier.Pop()
		if item == nil {
			t.Fatalf("Pop() returned nil at index %d", i)
		}
		if item.URL != expected {
			t.Errorf("Pop() order wrong at index %d: got %s, want %s", i, item.URL, expected)
		}
	}
}

func TestLinkPriorityConstants(t *testing.T) {
	// Verify priority ordering
	if LinkPrioritySidebar >= LinkPriorityContent {
		t.Errorf("Sidebar priority (%d) should be less than Content priority (%d)", 
			LinkPrioritySidebar, LinkPriorityContent)
	}
	if LinkPriorityContent >= LinkPriorityFooter {
		t.Errorf("Content priority (%d) should be less than Footer priority (%d)", 
			LinkPriorityContent, LinkPriorityFooter)
	}
}
