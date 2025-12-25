import os
import asyncio
from lightrag import LightRAG, QueryParam
from lightrag.llm.openai import gpt_4o_mini_complete, gpt_4o_complete, openai_embed
from lightrag.utils import setup_logger
from lightrag.utils import TokenTracker
from dotenv import load_dotenv

load_dotenv()

# Create TokenTracker instance
token_tracker = TokenTracker()

setup_logger("lightrag", level="INFO")

async def initialize_rag():
    rag = LightRAG(
        embedding_func=openai_embed,
        llm_model_func=gpt_4o_mini_complete,
        kv_storage="PGKVStorage",
        vector_storage="PGVectorStorage",
        graph_storage="PGGraphStorage",
        doc_status_storage="PGDocStatusStorage",
        workspace="unique_workspace_name_2",
    )
    # IMPORTANT: Both initialization calls are required!
    await rag.initialize_storages()  # Initialize storage backends
    return rag

async def main():
    try:
        # Initialize RAG instance
        rag = await initialize_rag()
        await rag.ainsert("python programming language", "Python is a high-level, interpreted programming language known for its readability and versatility.")
        await rag.ainsert("java programming language", "Java is a class-based, object-oriented programming language designed to have as few implementation dependencies as possible.")
        await rag.ainsert("javascript programming language", "JavaScript is a versatile, high-level programming language primarily used for web development to create interactive effects within web browsers.")
        await rag.ainsert("go programming language", "Go, also known as Golang, is a statically typed, compiled programming language designed for simplicity and efficiency, particularly in concurrent programming and scalable systems.")

        # Perform hybrid search
        with token_tracker:
            mode = "hybrid"
            print(
            await rag.aquery(
                "What is Python?",
                param=QueryParam(mode=mode)
            )
            )

    except Exception as e:
        print(f"An error occurred: {e}")
    finally:
        if rag:
            await rag.finalize_storages()

if __name__ == "__main__":
    asyncio.run(main())