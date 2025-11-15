# src/docs_mcp/indexer/chunker.py

def chunk_text(text: str, max_tokens: int = 500):
    words = text.split()
    chunks = []

    for i in range(0, len(words), max_tokens):
        chunk = " ".join(words[i:i + max_tokens])
        if len(chunk.strip()) > 0:
            chunks.append(chunk)

    return chunks
