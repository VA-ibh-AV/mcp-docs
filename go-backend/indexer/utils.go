package indexer

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CompressHTML compresses HTML content using gzip and base64 encodes it
func CompressHTML(html string) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	
	if _, err := gz.Write([]byte(html)); err != nil {
		return "", fmt.Errorf("gzip write error: %w", err)
	}
	
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("gzip close error: %w", err)
	}
	
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DecompressHTML decompresses base64+gzip encoded HTML
func DecompressHTML(compressed string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(compressed)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %w", err)
	}
	
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("gzip reader error: %w", err)
	}
	defer gz.Close()
	
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(gz); err != nil {
		return "", fmt.Errorf("gzip read error: %w", err)
	}
	
	return buf.String(), nil
}

// CreatePageContent creates a PageContent struct from crawl result
func CreatePageContent(result *CrawlResult, compress bool) (*PageContent, error) {
	content := &PageContent{
		Text:        result.Text,
		Title:       result.Title,
		ContentType: result.ContentType,
		HTMLSize:    len(result.HTML),
	}
	
	if compress && len(result.HTML) > 0 {
		compressed, err := CompressHTML(result.HTML)
		if err != nil {
			return nil, fmt.Errorf("failed to compress HTML: %w", err)
		}
		content.HTML = compressed
		content.Encoding = "gzip+base64"
	} else {
		content.HTML = result.HTML
		content.Encoding = "plain"
	}
	
	return content, nil
}

// CreateKafkaMessage creates a Kafka message from crawl result
func CreateKafkaMessage(
	jobID, requestID, projectID uint,
	userID string,
	collectionID string,
	result *CrawlResult,
	metadata map[string]string,
	compress bool,
) (*IndexingJobMessage, error) {
	content, err := CreatePageContent(result, compress)
	if err != nil {
		return nil, err
	}
	
	return &IndexingJobMessage{
		JobID:        jobID,
		RequestID:    requestID,
		ProjectID:    projectID,
		UserID:       userID,
		CollectionID: collectionID,
		URL:          result.URL,
		Depth:        result.Depth,
		ParentURL:    result.ParentURL,
		Content:      content,
		DiscoveredAt: result.ProcessedAt,
		Metadata:     metadata,
	}, nil
}

// SerializeMessage serializes a Kafka message to JSON
func SerializeMessage(msg *IndexingJobMessage) (string, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize message: %w", err)
	}
	return string(data), nil
}

// DeserializeMessage deserializes a JSON message
func DeserializeMessage(data string) (*IndexingJobMessage, error) {
	var msg IndexingJobMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}
	return &msg, nil
}
