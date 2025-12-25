"""
Kafka consumer for processing indexing jobs.
Uses aiokafka for async consumption with consumer groups.
"""

import asyncio
import json
import logging
from typing import Callable, Optional

from aiokafka import AIOKafkaConsumer
from aiokafka.errors import KafkaError

from app.config import Settings
from models.messages import IndexingJobMessage

logger = logging.getLogger(__name__)


class IndexingJobConsumer:
    """
    Async Kafka consumer for indexing_jobs topic.
    Uses consumer groups for horizontal scaling.
    """
    
    def __init__(
        self,
        settings: Settings,
        message_handler: Callable[[IndexingJobMessage], None],
    ):
        self.settings = settings
        self.message_handler = message_handler
        self.consumer: Optional[AIOKafkaConsumer] = None
        self._running = False
        self._tasks: set[asyncio.Task] = set()
        
    async def start(self) -> None:
        """Initialize and start the Kafka consumer."""
        logger.info(
            "Starting Kafka consumer",
            extra={
                "bootstrap_servers": self.settings.kafka_bootstrap_servers,
                "group_id": self.settings.kafka_group_id,
                "topic": self.settings.kafka_topic,
            }
        )
        
        self.consumer = AIOKafkaConsumer(
            self.settings.kafka_topic,
            bootstrap_servers=self.settings.kafka_bootstrap_servers,
            group_id=self.settings.kafka_group_id,
            enable_auto_commit=self.settings.kafka_enable_auto_commit,
            auto_offset_reset=self.settings.kafka_auto_offset_reset,
            value_deserializer=lambda m: json.loads(m.decode("utf-8")),
            # Increase max message size for large HTML content
            max_partition_fetch_bytes=10 * 1024 * 1024,  # 10MB
        )
        
        await self.consumer.start()
        self._running = True
        logger.info("Kafka consumer started successfully")
        
    async def stop(self) -> None:
        """Gracefully stop the consumer."""
        logger.info("Stopping Kafka consumer...")
        self._running = False
        
        # Wait for in-flight tasks to complete
        if self._tasks:
            logger.info(f"Waiting for {len(self._tasks)} in-flight tasks to complete...")
            await asyncio.gather(*self._tasks, return_exceptions=True)
        
        if self.consumer:
            await self.consumer.stop()
            
        logger.info("Kafka consumer stopped")
        
    async def consume(self) -> None:
        """
        Main consumption loop.
        Processes messages and commits offsets after successful processing.
        """
        if not self.consumer:
            raise RuntimeError("Consumer not started. Call start() first.")
            
        logger.info("Starting message consumption loop...")
        
        try:
            async for message in self.consumer:
                if not self._running:
                    break
                    
                try:
                    # Parse message
                    job = IndexingJobMessage(**message.value)
                    
                    logger.info(
                        "Received indexing job",
                        extra={
                            "job_id": job.job_id,
                            "collection_id": job.collection_id,
                            "url": job.url,
                            "partition": message.partition,
                            "offset": message.offset,
                        }
                    )
                    
                    # Process message (handler is responsible for error handling)
                    await self.message_handler(job)
                    
                    # Commit offset after successful processing
                    await self.consumer.commit()
                    
                    logger.debug(
                        "Committed offset",
                        extra={
                            "partition": message.partition,
                            "offset": message.offset,
                        }
                    )
                    
                except json.JSONDecodeError as e:
                    logger.error(
                        "Failed to decode message",
                        extra={
                            "error": str(e),
                            "partition": message.partition,
                            "offset": message.offset,
                        }
                    )
                    # Commit to skip malformed message
                    await self.consumer.commit()
                    
                except Exception as e:
                    logger.exception(
                        "Error processing message",
                        extra={
                            "error": str(e),
                            "partition": message.partition,
                            "offset": message.offset,
                        }
                    )
                    # Don't commit - message will be reprocessed on restart
                    # In production, consider DLQ after max retries
                    
        except KafkaError as e:
            logger.error(f"Kafka error: {e}")
            raise
            
    async def run(self) -> None:
        """Start consumer and run until stopped."""
        await self.start()
        try:
            await self.consume()
        finally:
            await self.stop()
