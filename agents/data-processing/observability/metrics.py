"""
OpenTelemetry metrics initialization for Python RAG Agent.
"""
import logging
from typing import Optional

from opentelemetry import metrics
from opentelemetry.exporter.prometheus import PrometheusMetricReader
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_VERSION
from prometheus_client import start_http_server

logger = logging.getLogger(__name__)


class Metrics:
    """Metrics container for the RAG agent."""
    
    def __init__(self, meter: metrics.Meter):
        # Kafka metrics
        self.kafka_messages_consumed = meter.create_counter(
            name="kafka_messages_consumed_total",
            description="Total number of Kafka messages consumed",
            unit="1"
        )
        
        self.kafka_messages_processed = meter.create_counter(
            name="kafka_messages_processed_total",
            description="Total number of Kafka messages successfully processed",
            unit="1"
        )
        
        self.kafka_processing_errors = meter.create_counter(
            name="kafka_processing_errors_total",
            description="Total number of message processing errors",
            unit="1"
        )
        
        self.kafka_consumer_lag = meter.create_up_down_counter(
            name="kafka_consumer_lag",
            description="Kafka consumer lag (messages behind)",
            unit="1"
        )
        
        self.kafka_processing_duration = meter.create_histogram(
            name="kafka_processing_duration_seconds",
            description="Time taken to process a Kafka message",
            unit="s"
        )
        
        # RAG processing metrics
        self.rag_instances_active = meter.create_up_down_counter(
            name="rag_instances_active",
            description="Number of active RAG instances",
            unit="1"
        )
        
        self.rag_instances_total = meter.create_counter(
            name="rag_instances_total",
            description="Total number of RAG instances created",
            unit="1"
        )
        
        self.rag_documents_processed = meter.create_counter(
            name="rag_documents_processed_total",
            description="Total number of documents processed by RAG",
            unit="1"
        )
        
        self.rag_processing_duration = meter.create_histogram(
            name="rag_processing_duration_seconds",
            description="Time taken for RAG processing",
            unit="s"
        )
        
        # Document processing metrics
        self.html_documents_cleaned = meter.create_counter(
            name="html_documents_cleaned_total",
            description="Total number of HTML documents cleaned",
            unit="1"
        )
        
        self.html_cleaning_duration = meter.create_histogram(
            name="html_cleaning_duration_seconds",
            description="Time taken for HTML cleaning",
            unit="s"
        )
        
        self.document_chunks_created = meter.create_counter(
            name="document_chunks_created_total",
            description="Total number of document chunks created",
            unit="1"
        )
        
        # Indexing job metrics
        self.indexing_jobs_started = meter.create_counter(
            name="indexing_jobs_started_total",
            description="Total number of indexing jobs started",
            unit="1"
        )
        
        self.indexing_jobs_completed = meter.create_counter(
            name="indexing_jobs_completed_total",
            description="Total number of indexing jobs completed",
            unit="1"
        )
        
        self.indexing_jobs_failed = meter.create_counter(
            name="indexing_jobs_failed_total",
            description="Total number of indexing jobs failed",
            unit="1"
        )
        
        self.indexing_pages_total = meter.create_counter(
            name="indexing_pages_total",
            description="Total number of pages indexed",
            unit="1"
        )
        
        # API call metrics (to Go backend)
        self.api_calls_total = meter.create_counter(
            name="api_calls_total",
            description="Total number of API calls made",
            unit="1"
        )
        
        self.api_call_duration = meter.create_histogram(
            name="api_call_duration_seconds",
            description="Duration of API calls",
            unit="s"
        )
        
        self.api_call_errors = meter.create_counter(
            name="api_call_errors_total",
            description="Total number of API call errors",
            unit="1"
        )


def init_metrics(service_name: str = "mcp-docs-rag-agent", service_version: str = "1.0.0", port: int = 9091) -> Optional[Metrics]:
    """
    Initialize OpenTelemetry metrics with Prometheus exporter.
    
    Args:
        service_name: Name of the service
        service_version: Version of the service
        port: Port to expose Prometheus metrics on
        
    Returns:
        Metrics instance or None if initialization fails
    """
    try:
        # Create resource with service information
        resource = Resource(attributes={
            SERVICE_NAME: service_name,
            SERVICE_VERSION: service_version,
        })
        
        # Create Prometheus metric reader
        reader = PrometheusMetricReader()
        
        # Create meter provider
        provider = MeterProvider(
            resource=resource,
            metric_readers=[reader]
        )
        
        # Set global meter provider
        metrics.set_meter_provider(provider)
        
        # Create meter
        meter = metrics.get_meter(service_name)
        
        # Start Prometheus HTTP server
        start_http_server(port=port, addr="0.0.0.0")
        logger.info(f"Prometheus metrics server started on port {port}")
        
        # Create and return metrics instance
        metrics_instance = Metrics(meter)
        logger.info("OpenTelemetry metrics initialized successfully")
        
        return metrics_instance
        
    except Exception as e:
        logger.error(f"Failed to initialize metrics: {e}")
        return None
