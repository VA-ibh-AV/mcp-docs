"""
Main entry point for the Python RAG Agent.

This agent consumes indexing jobs from Kafka and processes them using LightRAG.
"""

import asyncio
import logging
import signal
import sys
from typing import Optional

from dotenv import load_dotenv

# Load environment variables first
load_dotenv()

from app.config import get_settings
from consumer.kafka_consumer import IndexingJobConsumer
from processor.pipeline import DocumentProcessor
from rag.pool import LightRAGPool
from observability import init_metrics, Metrics

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.StreamHandler(sys.stdout),
    ]
)

logger = logging.getLogger(__name__)


class RAGAgent:
    """
    Main RAG Agent application.
    
    Coordinates Kafka consumer, document processor, and LightRAG pool.
    """
    
    def __init__(self):
        self.settings = get_settings()
        self._running = False
        self._shutdown_event = asyncio.Event()
        
        # Initialize components
        self.metrics: Optional[Metrics] = None
        self.rag_pool: Optional[LightRAGPool] = None
        self.processor: Optional[DocumentProcessor] = None
        self.consumer: Optional[IndexingJobConsumer] = None
        
    async def initialize(self) -> None:
        """Initialize all components."""
        logger.info("Initializing RAG Agent...")
        
        # Configure logging level
        logging.getLogger().setLevel(self.settings.log_level)
        
        # Initialize OpenTelemetry metrics
        metrics_port = getattr(self.settings, 'metrics_port', 9091)
        self.metrics = init_metrics(
            service_name="mcp-docs-rag-agent",
            service_version="1.0.0",
            port=metrics_port
        )
        
        # Initialize LightRAG pool
        self.rag_pool = LightRAGPool(self.settings)
        
        # Initialize document processor (no DB status updates - tracked in memory)
        self.processor = DocumentProcessor(
            settings=self.settings,
            rag_pool=self.rag_pool,
            metrics=self.metrics,
        )
        
        # Initialize Kafka consumer with processor as message handler
        self.consumer = IndexingJobConsumer(
            settings=self.settings,
            message_handler=self.processor.process,
            metrics=self.metrics,
        )
        
        logger.info("RAG Agent initialized successfully")
        logger.info(f"Configuration:")
        logger.info(f"  Kafka: {self.settings.kafka_bootstrap_servers}")
        logger.info(f"  Topic: {self.settings.kafka_topic}")
        logger.info(f"  Consumer Group: {self.settings.kafka_group_id}")
        logger.info(f"  Max Workers: {self.settings.max_workers}")
        logger.info(f"  Max RAG Instances: {self.settings.max_rag_instances}")
        logger.info(f"  Metrics Port: {metrics_port}")
        
    async def run(self) -> None:
        """Run the agent until shutdown."""
        self._running = True
        
        logger.info("Starting RAG Agent...")
        
        # Start cleanup task
        cleanup_task = asyncio.create_task(self._cleanup_loop())
        
        try:
            # Run consumer
            await self.consumer.run()
        except asyncio.CancelledError:
            logger.info("RAG Agent cancelled")
        except Exception as e:
            logger.exception(f"RAG Agent error: {e}")
        finally:
            self._running = False
            cleanup_task.cancel()
            
    async def shutdown(self) -> None:
        """Gracefully shutdown the agent."""
        logger.info("Shutting down RAG Agent...")
        self._running = False
        
        # Stop consumer
        if self.consumer:
            await self.consumer.stop()
        
        # Close all RAG instances
        if self.rag_pool:
            await self.rag_pool.close_all()
        
        logger.info("RAG Agent shutdown complete")
        
    async def _cleanup_loop(self) -> None:
        """Periodically cleanup expired RAG instances."""
        while self._running:
            try:
                await asyncio.sleep(60)  # Check every minute
                
                if self.rag_pool:
                    cleaned = await self.rag_pool.cleanup_expired()
                    if cleaned > 0:
                        logger.info(f"Cleaned up {cleaned} expired RAG instances")
                        
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in cleanup loop: {e}")


async def main():
    """Main entry point."""
    agent = RAGAgent()
    
    # Setup signal handlers
    loop = asyncio.get_event_loop()
    
    def handle_signal():
        logger.info("Received shutdown signal")
        asyncio.create_task(agent.shutdown())
    
    for sig in (signal.SIGTERM, signal.SIGINT):
        loop.add_signal_handler(sig, handle_signal)
    
    try:
        await agent.initialize()
        await agent.run()
    except KeyboardInterrupt:
        logger.info("Interrupted by user")
    finally:
        await agent.shutdown()


if __name__ == "__main__":
    asyncio.run(main())
