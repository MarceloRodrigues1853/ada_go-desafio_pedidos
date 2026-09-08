"""Demonstração interativa da orquestração RAG com LangGraph."""

from __future__ import annotations

import sys

from demo_rag import PROJECT_ROOT, load_api_key
from langgraph_rag import build_rag_graph
from rag_tutor import DEFAULT_MIN_SIMILARITY, GeminiEmbedder, LocalRAG, discover_markdown_files


def main() -> int:
    try:
        question = input("Digite sua pergunta sobre a documentação: ")
    except (EOFError, KeyboardInterrupt):
        print("\nDemonstração cancelada.")
        return 1

    api_key = load_api_key()
    if not api_key and question.strip():
        print("Erro: GOOGLE_API_KEY não foi encontrada no ambiente nem no arquivo .env.")
        return 1

    executed_nodes: list[str] = []
    rag = LocalRAG(PROJECT_ROOT, GeminiEmbedder(api_key)) if api_key else LocalRAG(PROJECT_ROOT, lambda *_: [])
    document_count = len(discover_markdown_files(PROJECT_ROOT))

    try:
        chunk_count = rag.build() if question.strip() else 0
        graph = build_rag_graph(rag, executed_nodes.append)
        final_state = graph.invoke(
            {
                "question": question,
                "threshold": DEFAULT_MIN_SIMILARITY,
                "document_count": document_count,
                "chunk_count": chunk_count,
            }
        )
    except RuntimeError as error:
        print(f"Erro ao preparar embeddings do Gemini: {error}")
        return 1

    print("\nSequência de nós:")
    print("START → " + " → ".join(executed_nodes) + " → END")
    print("\nEstado final relevante:")
    print(f"Threshold: {final_state['threshold']:.2f}")
    print(f"Documentos indexados: {final_state['document_count']}")
    print(f"Trechos indexados: {final_state['chunk_count']}")
    print(f"Evidência suficiente: {'sim' if final_state['has_evidence'] else 'não'}")
    if final_state.get("error"):
        print(f"Erro: {final_state['error']}")
    print("\nResposta:")
    print(final_state["answer"])
    return 0


if __name__ == "__main__":
    sys.exit(main())
