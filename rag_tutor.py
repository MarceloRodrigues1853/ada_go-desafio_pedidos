"""RAG local e pequeno para a documentacao do projeto.

Somente a criacao dos embeddings usa a API do Gemini. O indice e a busca por
similaridade de cosseno permanecem locais e em memoria.
"""

from __future__ import annotations

import json
import math
from dataclasses import dataclass
from pathlib import Path
from typing import Callable, Sequence
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


EmbeddingFunction = Callable[[str, str], Sequence[float]]
NO_EVIDENCE = "Não encontrei evidência suficiente na documentação indexada."


@dataclass(frozen=True)
class Chunk:
    source: str
    text: str
    position: int


@dataclass(frozen=True)
class SearchResult:
    chunk: Chunk
    score: float


def _is_within(path: Path, root: Path) -> bool:
    try:
        path.relative_to(root)
        return True
    except ValueError:
        return False


def safe_markdown_path(root: Path, candidate: Path) -> Path:
    """Valida um Markdown permitido e impede escape da raiz por path ou symlink."""
    root = root.resolve()
    resolved = candidate.resolve()
    if not _is_within(resolved, root):
        raise ValueError("acesso fora da raiz do projeto")

    relative = resolved.relative_to(root)
    lowered = {part.lower() for part in relative.parts}
    if ".git" in lowered or any(part == ".env" or part.startswith(".env.") for part in lowered):
        raise ValueError("arquivo protegido")
    if resolved.suffix.lower() != ".md":
        raise ValueError("somente arquivos Markdown são permitidos")

    top_level = relative.as_posix() in {"README.md", "DIAGNOSTICO.md"}
    inside_docs = len(relative.parts) > 1 and relative.parts[0].lower() == "docs"
    if not (top_level or inside_docs):
        raise ValueError("arquivo fora da allowlist de documentação")
    return resolved


def discover_markdown_files(root: Path) -> list[Path]:
    """Descobre apenas README, DIAGNOSTICO e Markdown sob docs/."""
    root = root.resolve()
    candidates = [root / "README.md", root / "DIAGNOSTICO.md"]
    docs = root / "docs"
    if docs.is_dir():
        candidates.extend(docs.rglob("*.md"))

    files: list[Path] = []
    for candidate in candidates:
        if not candidate.is_file():
            continue
        try:
            files.append(safe_markdown_path(root, candidate))
        except ValueError:
            continue
    return sorted(set(files), key=lambda path: path.relative_to(root).as_posix())


def split_markdown(text: str, source: str, chunk_size: int = 900, overlap: int = 120) -> list[Chunk]:
    """Divide texto por palavras, preservando uma pequena sobreposição."""
    if chunk_size <= 0 or overlap < 0 or overlap >= chunk_size:
        raise ValueError("chunk_size deve ser positivo e overlap menor que chunk_size")

    words = text.split()
    chunks: list[Chunk] = []
    start = 0
    position = 0
    while start < len(words):
        end = start
        length = 0
        while end < len(words):
            addition = len(words[end]) + (1 if end > start else 0)
            if end > start and length + addition > chunk_size:
                break
            length += addition
            end += 1
        chunks.append(Chunk(source=source, text=" ".join(words[start:end]), position=position))
        position += 1
        if end == len(words):
            break
        next_start = end
        retained = 0
        while next_start > start and retained < overlap:
            next_start -= 1
            retained += len(words[next_start]) + 1
        start = max(start + 1, next_start)
    return chunks


def cosine_similarity(left: Sequence[float], right: Sequence[float]) -> float:
    if len(left) != len(right) or not left:
        raise ValueError("embeddings devem ter a mesma dimensão e não podem ser vazios")
    denominator = math.sqrt(sum(value * value for value in left)) * math.sqrt(
        sum(value * value for value in right)
    )
    if denominator == 0:
        return 0.0
    return sum(a * b for a, b in zip(left, right)) / denominator


class GeminiEmbedder:
    """Cliente REST mínimo para embeddings Gemini, sem SDK adicional."""

    def __init__(self, api_key: str, model: str = "gemini-embedding-001", timeout: int = 30):
        if not api_key:
            raise ValueError("GOOGLE_API_KEY não configurada")
        self.api_key = api_key
        self.model = model
        self.timeout = timeout

    def __call__(self, text: str, task_type: str) -> list[float]:
        url = f"https://generativelanguage.googleapis.com/v1beta/models/{self.model}:embedContent"
        body = json.dumps(
            {"content": {"parts": [{"text": text}]}, "taskType": task_type}
        ).encode("utf-8")
        request = Request(
            url,
            data=body,
            method="POST",
            headers={"Content-Type": "application/json", "x-goog-api-key": self.api_key},
        )
        try:
            with urlopen(request, timeout=self.timeout) as response:
                payload = json.load(response)
        except HTTPError as error:
            detail = error.read().decode("utf-8", errors="replace")[:500]
            raise RuntimeError(f"Gemini embeddings retornou HTTP {error.code}: {detail}") from error
        except URLError as error:
            raise RuntimeError(f"falha ao acessar Gemini embeddings: {error.reason}") from error
        try:
            return [float(value) for value in payload["embedding"]["values"]]
        except (KeyError, TypeError, ValueError) as error:
            raise RuntimeError("resposta inválida da API de embeddings") from error


class LocalRAG:
    def __init__(self, root: Path, embed: EmbeddingFunction):
        self.root = root.resolve()
        self.embed = embed
        self._index: list[tuple[Chunk, Sequence[float]]] = []

    def build(self) -> int:
        self._index.clear()
        for path in discover_markdown_files(self.root):
            source = path.relative_to(self.root).as_posix()
            text = path.read_text(encoding="utf-8", errors="replace")
            for chunk in split_markdown(text, source):
                prepared = f"title: {source} | text: {chunk.text}"
                vector = self.embed(prepared, "RETRIEVAL_DOCUMENT")
                self._index.append((chunk, vector))
        return len(self._index)

    def search(self, query: str, limit: int = 3, min_score: float = 0.55) -> list[SearchResult]:
        if not query.strip() or not self._index:
            return []
        query_vector = self.embed(query, "RETRIEVAL_QUERY")
        results = [
            SearchResult(chunk=chunk, score=cosine_similarity(query_vector, vector))
            for chunk, vector in self._index
        ]
        results.sort(key=lambda result: result.score, reverse=True)
        return [result for result in results if result.score >= min_score][:limit]

    def context(self, query: str, limit: int = 3, min_score: float = 0.55) -> str:
        results = self.search(query, limit=limit, min_score=min_score)
        if not results:
            return NO_EVIDENCE
        blocks = []
        for result in results:
            blocks.append(
                f"Fonte: {result.chunk.source} (trecho {result.chunk.position + 1}, "
                f"similaridade {result.score:.3f})\n{result.chunk.text}"
            )
        return "\n\n---\n\n".join(blocks)
