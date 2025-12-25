package observability

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsMiddleware creates a Gin middleware for recording HTTP metrics
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		ctx := context.Background()
		
		// Track active requests
		metrics.HTTPActiveRequests.Add(ctx, 1)
		defer metrics.HTTPActiveRequests.Add(ctx, -1)

		// Track request size
		if c.Request.ContentLength > 0 {
			attrs := []attribute.KeyValue{
				attribute.String("method", c.Request.Method),
				attribute.String("path", path),
			}
			metrics.HTTPRequestSize.Record(ctx, c.Request.ContentLength, metric.WithAttributes(attrs...))
		}

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method
		responseSize := int64(c.Writer.Size())

		attrs := []attribute.KeyValue{
			attribute.String("method", method),
			attribute.String("path", path),
			attribute.Int("status", status),
			attribute.String("status_class", getStatusClass(status)),
		}

		// Increment request counter
		metrics.HTTPRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

		// Record request duration
		metrics.HTTPRequestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))

		// Record response size
		if responseSize > 0 {
			metrics.HTTPResponseSize.Record(ctx, responseSize, metric.WithAttributes(attrs...))
		}
	}
}

// getStatusClass returns the status class (1xx, 2xx, 3xx, 4xx, 5xx)
func getStatusClass(status int) string {
	switch {
	case status >= 100 && status < 200:
		return "1xx"
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
