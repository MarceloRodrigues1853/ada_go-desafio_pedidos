"""Demonstração interativa do RAG local do agente tutor."""

from __future__ import annotations

import os
from pathlib import Path
from typing import Callable, Sequence, TextIO

from rag_tutor import (
    DEFAULT_MIN_SIMILARITY,
    GeminiEmbedder,
    LocalRAG,
    NO_EVIDENCE,
    discover_markdown_files,
)


PROJECT_ROOT = Path(__file__).resolve().parent


def load_api_key(root: Path = PROJECT_ROOT) -> str | None:
    """Obtém a chave do ambiente ou somente sua entrada específica no .env."""
    configured = os.getenv("GOOGLE_API_KEY")
    if configured:
        return configured

    env_file = root / ".env"
    if not env_file.is_file():
        return None
    for raw_line in env_file.read_text(encoding="utf-8-sig").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[7:].lstrip()
        name, separator, value = line.partition("=")
        if separator and name.strip() == "GOOGLE_API_KEY":
            return value.strip().strip('"\'') or None
    return None


def run_demo(
    embed: Callable[[str, str], Sequence[float]],
    question: str,
    output: TextIO,
    root: Path = PROJECT_ROOT,
    min_score: float = DEFAULT_MIN_SIMILARITY,
) -> int:
    """Indexa a documentação e imprime até três evidências para uma pergunta."""
    print(f"Limite mínimo de similaridade: {min_score:.2f}", file=output)
    if not question.strip():
        print("Pergunta vazia; a busca não foi executada.", file=output)
        print(NO_EVIDENCE, file=output)
        return 0

    document_count = len(discover_markdown_files(root))
    rag = LocalRAG(root, embed)
    chunk_count = rag.build()

    print(f"Documentos indexados: {document_count}", file=output)
    print(f"Trechos indexados: {chunk_count}", file=output)

    results = rag.search(question, limit=3, min_score=min_score)
    if not results:
        print("\nNenhuma evidência suficiente.", file=output)
        print(NO_EVIDENCE, file=output)
        return 0

    print(f"\nTrechos recuperados: {len(results)}", file=output)
    for number, result in enumerate(results, start=1):
        print(f"\n--- Resultado {number} ---", file=output)
        print(f"Arquivo: {result.chunk.source}", file=output)
        print(f"Posição: {result.chunk.position + 1}", file=output)
        print(f"Similaridade: {result.score:.3f}", file=output)
        print("Conteúdo:", file=output)
        print(result.chunk.text, file=output)
    return 0


def main() -> int:
    try:
        question = input("Digite sua pergunta sobre a documentação: ").strip()
    except (EOFError, KeyboardInterrupt):
        print("\nDemonstração cancelada.")
        return 1

    if not question:
        print("Pergunta vazia; a busca não foi executada.")
        print(NO_EVIDENCE)
        return 0

    api_key = load_api_key()
    if not api_key:
        print(
            "Erro: GOOGLE_API_KEY não foi encontrada no ambiente nem no arquivo .env."
        )
        return 1

    try:
        return run_demo(GeminiEmbedder(api_key), question, output=os.sys.stdout)
    except RuntimeError as error:
        print(f"Erro ao consultar embeddings do Gemini: {error}")
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
