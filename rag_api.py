"""API HTTP mínima e somente leitura para o RAG documental do projeto."""

from __future__ import annotations

import json
import os
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from typing import Any

from rag_tutor import (
    DEFAULT_MIN_SIMILARITY,
    GeminiEmbedder,
    LocalRAG,
    NO_EVIDENCE,
    discover_markdown_files,
)


PROJECT_ROOT = Path(__file__).resolve().parent
DEFAULT_PORT = 8081
MAX_BODY_BYTES = 4096
MAX_QUESTION_CHARS = 500


def parse_allowed_origins(configured: str) -> frozenset[str]:
    return frozenset(origin.strip() for origin in configured.split(",") if origin.strip())


class RAGQueryService:
    """Indexa uma vez por processo e executa o grafo para cada pergunta."""

    def __init__(self, rag: LocalRAG, root: Path):
        self.rag = rag
        self.root = root.resolve()
        self.document_count = len(discover_markdown_files(self.root))
        self.chunk_count = 0
        self._indexed = False
        self._index_lock = threading.Lock()

    def _ensure_index(self) -> None:
        if self._indexed:
            return
        with self._index_lock:
            if not self._indexed:
                self.chunk_count = self.rag.build()
                self._indexed = True

    def ask(self, question: str) -> dict[str, Any]:
        normalized = question.strip()
        if not normalized:
            return self._response(NO_EVIDENCE, False, [])
        if len(normalized) > MAX_QUESTION_CHARS:
            raise ValueError(f"a pergunta deve ter no máximo {MAX_QUESTION_CHARS} caracteres")

        self._ensure_index()
        results = self.rag.search(
            normalized, limit=3, min_score=DEFAULT_MIN_SIMILARITY
        )
        sources = [
            {
                "file": result.chunk.source,
                "position": result.chunk.position + 1,
                "similarity": round(result.score, 3),
                "content": result.chunk.text,
            }
            for result in results
        ]
        if not results:
            return self._response(NO_EVIDENCE, False, [])
        answer = "\n\n---\n\n".join(
            f"Fonte: {result.chunk.source} (trecho {result.chunk.position + 1}, "
            f"similaridade {result.score:.3f})\n{result.chunk.text}"
            for result in results
        )
        return self._response(answer, True, sources)

    def _response(
        self, answer: str, has_evidence: bool, sources: list[dict[str, Any]]
    ) -> dict[str, Any]:
        return {
            "answer": answer,
            "has_evidence": has_evidence,
            "sources": sources,
            "threshold": DEFAULT_MIN_SIMILARITY,
            "documents_indexed": self.document_count,
            "chunks_indexed": self.chunk_count,
        }


def make_handler(service: RAGQueryService, allowed_origins: frozenset[str]):
    class RAGRequestHandler(BaseHTTPRequestHandler):
        def _send_json(self, status: int, payload: dict[str, Any]) -> None:
            encoded = json.dumps(payload, ensure_ascii=False).encode("utf-8")
            self.send_response(status)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Content-Length", str(len(encoded)))
            origin = self.headers.get("Origin", "")
            if origin in allowed_origins:
                self.send_header("Access-Control-Allow-Origin", origin)
                self.send_header("Vary", "Origin")
            self.end_headers()
            self.wfile.write(encoded)

        def _origin_is_allowed(self) -> bool:
            origin = self.headers.get("Origin")
            return origin is None or origin in allowed_origins

        def do_OPTIONS(self) -> None:
            if not self._origin_is_allowed():
                self._send_json(403, {"error": "origem não permitida"})
                return
            self.send_response(204)
            origin = self.headers.get("Origin", "")
            if origin:
                self.send_header("Access-Control-Allow-Origin", origin)
                self.send_header("Vary", "Origin")
            self.send_header("Access-Control-Allow-Headers", "Content-Type")
            self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            self.end_headers()

        def do_GET(self) -> None:
            if self.path != "/health":
                self._send_json(404, {"error": "rota não encontrada"})
                return
            self._send_json(200, {"status": "ok"})

        def do_POST(self) -> None:
            if self.path != "/ask":
                self._send_json(404, {"error": "rota não encontrada"})
                return
            if not self._origin_is_allowed():
                self._send_json(403, {"error": "origem não permitida"})
                return
            try:
                content_length = int(self.headers.get("Content-Length", "0"))
                if content_length <= 0 or content_length > MAX_BODY_BYTES:
                    raise ValueError("corpo da requisição inválido")
                payload = json.loads(self.rfile.read(content_length))
                question = payload.get("question", "") if isinstance(payload, dict) else ""
                if not isinstance(question, str):
                    raise ValueError("question deve ser texto")
                self._send_json(200, service.ask(question))
            except (json.JSONDecodeError, ValueError) as error:
                self._send_json(400, {"error": str(error)})
            except Exception:
                self._send_json(503, {"error": "não foi possível consultar o RAG"})

        def log_message(self, format: str, *args: object) -> None:
            return

    return RAGRequestHandler


def main() -> int:
    api_key = os.getenv("GOOGLE_API_KEY", "")
    if not api_key:
        raise RuntimeError("GOOGLE_API_KEY não configurada no ambiente")

    allowed_origins = parse_allowed_origins(os.getenv("RAG_CORS_ALLOWED_ORIGINS", ""))
    port = int(os.getenv("PORT", str(DEFAULT_PORT)))
    service = RAGQueryService(LocalRAG(PROJECT_ROOT, GeminiEmbedder(api_key)), PROJECT_ROOT)
    server = ThreadingHTTPServer(("0.0.0.0", port), make_handler(service, allowed_origins))
    print(f"RAG API ativa na porta {port}")
    server.serve_forever()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
